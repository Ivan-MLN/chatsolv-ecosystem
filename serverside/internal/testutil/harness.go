package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"authbackend/internal/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
)

const dependencyTimeout = 10 * time.Second

// Postgres returns a pool backed by a unique migrated schema. The schema is
// dropped automatically when the test finishes.
func Postgres(t testing.TB) *pgxpool.Pool {
	t.Helper()
	url := dependencyURL(t, "TEST_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/chatsolv_test?sslmode=disable")
	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	ctx, cancel := context.WithTimeout(context.Background(), dependencyTimeout)
	defer cancel()

	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		dependencyUnavailable(t, "PostgreSQL", err)
	}
	if err = admin.Ping(ctx); err != nil {
		admin.Close()
		dependencyUnavailable(t, "PostgreSQL", err)
	}
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+pgx.Identifier{schema}.Sanitize()); err != nil {
		admin.Close()
		t.Fatalf("create PostgreSQL test schema: %v", err)
	}
	if err = applyMigrations(ctx, admin, schema); err != nil {
		_, _ = admin.Exec(context.Background(), "DROP SCHEMA "+pgx.Identifier{schema}.Sanitize()+" CASCADE")
		admin.Close()
		t.Fatalf("apply PostgreSQL migrations: %v", err)
	}

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatalf("parse PostgreSQL test URL: %v", err)
	}
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET search_path TO "+pgx.Identifier{schema}.Sanitize())
		return err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("create isolated PostgreSQL pool: %v", err)
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping isolated PostgreSQL pool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), dependencyTimeout)
		defer cleanupCancel()
		if _, cleanupErr := admin.Exec(cleanupCtx, "DROP SCHEMA "+pgx.Identifier{schema}.Sanitize()+" CASCADE"); cleanupErr != nil {
			t.Errorf("drop PostgreSQL test schema: %v", cleanupErr)
		}
		admin.Close()
	})
	return pool
}

type RedisHarness struct {
	Client *redis.Client
	prefix string
}

func (h *RedisHarness) Key(value string) string { return h.prefix + value }

// Redis returns a shared client with a test-specific key prefix and cleanup.
func Redis(t testing.TB) *RedisHarness {
	t.Helper()
	url := dependencyURL(t, "TEST_REDIS_URL", "redis://localhost:6379/15")
	options, err := redis.ParseURL(url)
	if err != nil {
		t.Fatalf("parse Redis test URL: %v", err)
	}
	client := redis.NewClient(options)
	ctx, cancel := context.WithTimeout(context.Background(), dependencyTimeout)
	defer cancel()
	if err = client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		dependencyUnavailable(t, "Redis", err)
	}
	harness := &RedisHarness{Client: client, prefix: "test:" + uuid.NewString() + ":"}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), dependencyTimeout)
		defer cleanupCancel()
		var cursor uint64
		for {
			keys, next, scanErr := client.Scan(cleanupCtx, cursor, harness.prefix+"*", 100).Result()
			if scanErr != nil {
				t.Errorf("scan Redis test keys: %v", scanErr)
				break
			}
			if len(keys) > 0 {
				if deleteErr := client.Del(cleanupCtx, keys...).Err(); deleteErr != nil {
					t.Errorf("delete Redis test keys: %v", deleteErr)
				}
			}
			cursor = next
			if cursor == 0 {
				break
			}
		}
		_ = client.Close()
	})
	return harness
}

type MinIOHarness struct {
	Store *storage.MinIO
}

// MinIO returns a store backed by a unique bucket removed after the test.
func MinIO(t testing.TB) *MinIOHarness {
	t.Helper()
	endpoint := dependencyURL(t, "TEST_MINIO_ENDPOINT", "localhost:9000")
	accessKey := envOr("TEST_MINIO_ACCESS_KEY", "chatsolv")
	secretKey := envOr("TEST_MINIO_SECRET_KEY", "chatsolv-development-secret")
	bucket := "test-" + strings.ReplaceAll(uuid.NewString(), "-", "")
	store, err := storage.NewMinIO(endpoint, accessKey, secretKey, bucket, false)
	if err != nil {
		t.Fatalf("create MinIO test client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), dependencyTimeout)
	defer cancel()
	if err = store.EnsureBucket(ctx); err != nil {
		dependencyUnavailable(t, "MinIO", err)
	}

	client, err := minio.New(strings.TrimPrefix(strings.TrimPrefix(endpoint, "http://"), "https://"), &minio.Options{Creds: minioCredentials(accessKey, secretKey)})
	if err != nil {
		t.Fatalf("create MinIO cleanup client: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), dependencyTimeout)
		defer cleanupCancel()
		for object := range client.ListObjects(cleanupCtx, bucket, minio.ListObjectsOptions{Recursive: true}) {
			if object.Err != nil {
				t.Errorf("list MinIO test objects: %v", object.Err)
				continue
			}
			if removeErr := client.RemoveObject(cleanupCtx, bucket, object.Key, minio.RemoveObjectOptions{}); removeErr != nil {
				t.Errorf("remove MinIO test object: %v", removeErr)
			}
		}
		if removeErr := client.RemoveBucket(cleanupCtx, bucket); removeErr != nil {
			t.Errorf("remove MinIO test bucket: %v", removeErr)
		}
	})
	return &MinIOHarness{Store: store}
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool, schema string) error {
	entries, err := os.ReadDir(migrationsDir())
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err = tx.Exec(ctx, "SET LOCAL search_path TO "+pgx.Identifier{schema}.Sanitize()); err != nil {
		return err
	}
	for _, name := range names {
		migration, readErr := os.ReadFile(filepath.Join(migrationsDir(), name))
		if readErr != nil {
			return readErr
		}
		if _, err = tx.Exec(ctx, string(migration)); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return tx.Commit(ctx)
}

func migrationsDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "db", "migrations")
}

func dependencyURL(t testing.TB, key, fallback string) string {
	t.Helper()
	if value := os.Getenv(key); value != "" {
		return value
	}
	if integrationRequired() {
		return fallback
	}
	t.Skipf("%s not set; set TEST_INTEGRATION=1 to require local integration dependencies", key)
	return ""
}

func dependencyUnavailable(t testing.TB, name string, err error) {
	t.Helper()
	if integrationRequired() {
		t.Fatalf("%s integration dependency unavailable: %v", name, err)
	}
	t.Skipf("%s integration dependency unavailable: %v", name, err)
}

func integrationRequired() bool { return os.Getenv("TEST_INTEGRATION") == "1" }

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func minioCredentials(accessKey, secretKey string) *credentials.Credentials {
	return credentials.NewStaticV4(accessKey, secretKey, "")
}
