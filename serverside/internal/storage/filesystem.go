package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("object not found")
var workspacePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type Object struct {
	Key, Checksum string
	Size          int64
}
type Store interface {
	Put(context.Context, string, string, io.Reader) (Object, error)
	Get(context.Context, string, string) (io.ReadCloser, error)
	Delete(context.Context, string, string) error
}
type Filesystem struct{ root string }

func NewFilesystem(root string) *Filesystem {
	absolute, _ := filepath.Abs(root)
	return &Filesystem{filepath.Clean(absolute)}
}
func (s *Filesystem) Put(_ context.Context, workspaceID, filename string, reader io.Reader) (Object, error) {
	if !workspacePattern.MatchString(workspaceID) {
		return Object{}, errors.New("invalid workspace")
	}
	dir := filepath.Join(s.root, workspaceID)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return Object{}, err
	}
	ext := strings.ToLower(filepath.Ext(filepath.Base(filename)))
	if len(ext) > 10 {
		ext = ""
	}
	key := workspaceID + "/" + uuid.NewString() + ext
	path := filepath.Join(s.root, filepath.FromSlash(key))
	if !within(s.root, path) {
		return Object{}, errors.New("unsafe object path")
	}
	tmp, err := os.CreateTemp(dir, ".upload-*")
	if err != nil {
		return Object{}, err
	}
	defer os.Remove(tmp.Name())
	hash := sha256.New()
	size, err := io.Copy(io.MultiWriter(tmp, hash), reader)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return Object{}, err
	}
	if err = os.Rename(tmp.Name(), path); err != nil {
		return Object{}, err
	}
	return Object{Key: key, Checksum: hex.EncodeToString(hash.Sum(nil)), Size: size}, nil
}
func (s *Filesystem) Get(_ context.Context, workspaceID, key string) (io.ReadCloser, error) {
	prefix := workspaceID + "/"
	if !workspacePattern.MatchString(workspaceID) || !strings.HasPrefix(key, prefix) {
		return nil, ErrNotFound
	}
	path := filepath.Join(s.root, filepath.FromSlash(key))
	if !within(filepath.Join(s.root, workspaceID), path) {
		return nil, ErrNotFound
	}
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	return file, err
}
func (s *Filesystem) Delete(_ context.Context, workspaceID, key string) error {
	r, err := s.Get(context.Background(), workspaceID, key)
	if err != nil {
		return err
	}
	_ = r.Close()
	path := filepath.Join(s.root, filepath.FromSlash(key))
	if err = os.Remove(path); err != nil {
		return fmt.Errorf("delete object: %w", err)
	}
	return nil
}
func within(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}
