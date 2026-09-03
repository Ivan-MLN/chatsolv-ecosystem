package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIO struct {
	client *minio.Client
	bucket string
}

func NewMinIO(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinIO, error) {
	client, err := minio.New(strings.TrimPrefix(strings.TrimPrefix(endpoint, "http://"), "https://"), &minio.Options{Creds: credentials.NewStaticV4(accessKey, secretKey, ""), Secure: useSSL})
	if err != nil {
		return nil, err
	}
	return &MinIO{client: client, bucket: bucket}, nil
}

func (s *MinIO) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
}

func (s *MinIO) Put(ctx context.Context, workspaceID, filename string, reader io.Reader) (Object, error) {
	if !workspacePattern.MatchString(workspaceID) {
		return Object{}, errors.New("invalid workspace")
	}
	ext := strings.ToLower(filepath.Ext(filepath.Base(filename)))
	if len(ext) > 10 {
		ext = ""
	}
	key := workspaceID + "/" + uuid.NewString() + ext
	hash := sha256.New()
	result, err := s.client.PutObject(ctx, s.bucket, key, io.TeeReader(reader, hash), -1, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return Object{}, err
	}
	return Object{Key: key, Checksum: hex.EncodeToString(hash.Sum(nil)), Size: result.Size}, nil
}

func (s *MinIO) Get(ctx context.Context, workspaceID, key string) (io.ReadCloser, error) {
	if !workspacePattern.MatchString(workspaceID) || !strings.HasPrefix(key, workspaceID+"/") {
		return nil, ErrNotFound
	}
	object, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	if _, err = object.Stat(); err != nil {
		_ = object.Close()
		return nil, ErrNotFound
	}
	return object, nil
}

func (s *MinIO) Delete(ctx context.Context, workspaceID, key string) error {
	if !workspacePattern.MatchString(workspaceID) || !strings.HasPrefix(key, workspaceID+"/") {
		return ErrNotFound
	}
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}
