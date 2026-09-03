package storage

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

func TestFilesystemScopesObjectsByWorkspaceAndRejectsTraversal(t *testing.T) {
	s := NewFilesystem(t.TempDir())
	object, err := s.Put(context.Background(), "wsp_a", "../../evil.pdf", bytes.NewBufferString("safe"))
	require.NoError(t, err)
	require.NotContains(t, object.Key, "..")
	r, err := s.Get(context.Background(), "wsp_b", object.Key)
	require.Error(t, err)
	require.Nil(t, r)
	r, err = s.Get(context.Background(), "wsp_a", object.Key)
	require.NoError(t, err)
	data, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Equal(t, "safe", string(data))
	require.NoError(t, r.Close())
}
func TestFilesystemCalculatesChecksum(t *testing.T) {
	s := NewFilesystem(t.TempDir())
	object, err := s.Put(context.Background(), "wsp_a", "catalog.txt", bytes.NewBufferString("hello"))
	require.NoError(t, err)
	require.Equal(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", object.Checksum)
	require.Equal(t, int64(5), object.Size)
}
