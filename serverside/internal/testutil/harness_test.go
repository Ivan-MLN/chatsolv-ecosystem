package testutil

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPostgresHarnessAppliesMigrationsAndIsolatesSchemas(t *testing.T) {
	first := Postgres(t)
	second := Postgres(t)
	ctx := context.Background()

	_, err := first.Exec(ctx, `INSERT INTO users (id, name, email, password_hash) VALUES ('00000000-0000-0000-0000-000000000001', 'Tenant A', 'a@example.com', 'hash')`)
	require.NoError(t, err)

	var firstCount, secondCount int
	require.NoError(t, first.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&firstCount))
	require.NoError(t, second.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&secondCount))
	require.Equal(t, 1, firstCount)
	require.Zero(t, secondCount)
}

func TestRedisHarnessUsesIsolatedKeyPrefixes(t *testing.T) {
	first := Redis(t)
	second := Redis(t)
	ctx := context.Background()

	require.NoError(t, first.Client.Set(ctx, first.Key("state"), "tenant-a", 0).Err())
	value, err := second.Client.Get(ctx, second.Key("state")).Result()
	require.Error(t, err)
	require.Empty(t, value)
}

func TestMinIOHarnessUsesIsolatedBuckets(t *testing.T) {
	first := MinIO(t)
	second := MinIO(t)
	ctx := context.Background()

	object, err := first.Store.Put(ctx, "wsp_a", "document.txt", bytes.NewBufferString("tenant-a"))
	require.NoError(t, err)

	reader, err := second.Store.Get(ctx, "wsp_a", object.Key)
	require.Error(t, err)
	require.Nil(t, reader)

	reader, err = first.Store.Get(ctx, "wsp_a", object.Key)
	require.NoError(t, err)
	defer reader.Close()
	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, "tenant-a", string(content))
}
