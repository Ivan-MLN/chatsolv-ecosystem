package obsidian

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateVaultIsIdempotentAndCreatesDefaultLayout(t *testing.T) {
	root := t.TempDir()
	brain := NewFilesystem(root)
	first, err := brain.CreateVault(context.Background(), "wsp_01tenant")
	require.NoError(t, err)
	second, err := brain.CreateVault(context.Background(), "wsp_01tenant")
	require.NoError(t, err)
	require.Equal(t, first, second)
	for _, path := range []string{".chatsolv", "business", "bot", "products", "services", "policies", "faq", "knowledge", "imports", "attachments"} {
		info, statErr := os.Stat(filepath.Join(first.Path, path))
		require.NoError(t, statErr)
		require.True(t, info.IsDir())
	}
}

func TestVaultRejectsTraversalAbsoluteAndSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	brain := NewFilesystem(root)
	vault, err := brain.CreateVault(context.Background(), "wsp_tenant")
	require.NoError(t, err)
	for _, path := range []string{"../../outside.md", "../other/file.md", "/tmp/absolute.md", "%2e%2e/outside.md"} {
		require.Error(t, brain.WriteNote(context.Background(), vault.ID, Note{Path: path, Content: "unsafe"}))
	}
	outside := t.TempDir()
	require.NoError(t, os.Symlink(outside, filepath.Join(vault.Path, "knowledge", "escape")))
	require.Error(t, brain.WriteNote(context.Background(), vault.ID, Note{Path: "knowledge/escape/note.md", Content: "unsafe"}))
}

func TestVaultIDResolvesInternallyAndCannotReadOtherTenant(t *testing.T) {
	brain := NewFilesystem(t.TempDir())
	a, err := brain.CreateVault(context.Background(), "wsp_a")
	require.NoError(t, err)
	b, err := brain.CreateVault(context.Background(), "wsp_b")
	require.NoError(t, err)
	require.NoError(t, brain.WriteNote(context.Background(), a.ID, Note{Path: "policies/refund.md", Content: "7 days"}))
	_, err = brain.ReadNote(context.Background(), b.ID, "../wsp_a/policies/refund.md")
	require.Error(t, err)
}
