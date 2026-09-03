package obsidian

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var ErrUnsafePath = errors.New("unsafe vault path")
var vaultIDPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

type Vault struct{ ID, Path string }
type Note struct{ Path, Content string }

type SecondBrain interface {
	CreateVault(context.Context, string) (Vault, error)
	WriteNote(context.Context, string, Note) error
	ReadNote(context.Context, string, string) (Note, error)
	DeleteNote(context.Context, string, string) error
	ListNotes(context.Context, string) ([]Note, error)
	DeleteVault(context.Context, string) error
}

type Filesystem struct{ root string }

func NewFilesystem(root string) *Filesystem {
	absolute, _ := filepath.Abs(root)
	return &Filesystem{root: filepath.Clean(absolute)}
}

func (f *Filesystem) VaultPath(vaultID string) (string, error) { return f.vaultPath(vaultID) }

func (f *Filesystem) CreateVault(_ context.Context, workspaceID string) (Vault, error) {
	if !vaultIDPattern.MatchString(workspaceID) {
		return Vault{}, ErrUnsafePath
	}
	path := filepath.Join(f.root, workspaceID)
	for _, dir := range []string{".chatsolv", "business", "bot", "products", "services", "policies", "faq", "knowledge", "imports", "attachments"} {
		if err := os.MkdirAll(filepath.Join(path, dir), 0o750); err != nil {
			return Vault{}, err
		}
	}
	return Vault{ID: workspaceID, Path: path}, nil
}

func (f *Filesystem) WriteNote(_ context.Context, vaultID string, note Note) error {
	path, err := f.safePath(vaultID, note.Path, true)
	if err != nil {
		return err
	}
	if filepath.Ext(path) != ".md" && !strings.HasSuffix(note.Path, ".json") {
		return ErrUnsafePath
	}
	if err = os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	if err = f.rejectSymlinkParents(vaultID, path); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".chatsolv-write-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err = tmp.Chmod(0o640); err == nil {
		_, err = tmp.WriteString(note.Content)
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func (f *Filesystem) ReadNote(_ context.Context, vaultID, relative string) (Note, error) {
	path, err := f.safePath(vaultID, relative, false)
	if err != nil {
		return Note{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Note{}, err
	}
	return Note{Path: relative, Content: string(data)}, nil
}
func (f *Filesystem) DeleteNote(_ context.Context, vaultID, relative string) error {
	path, err := f.safePath(vaultID, relative, false)
	if err != nil {
		return err
	}
	return os.Remove(path)
}
func (f *Filesystem) ListNotes(_ context.Context, vaultID string) ([]Note, error) {
	root, err := f.vaultPath(vaultID)
	if err != nil {
		return nil, err
	}
	var notes []Note
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return ErrUnsafePath
		}
		if entry.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		notes = append(notes, Note{Path: filepath.ToSlash(rel), Content: string(data)})
		return nil
	})
	sort.Slice(notes, func(i, j int) bool { return notes[i].Path < notes[j].Path })
	return notes, err
}
func (f *Filesystem) DeleteVault(_ context.Context, vaultID string) error {
	path, err := f.vaultPath(vaultID)
	if err != nil {
		return err
	}
	return os.RemoveAll(path)
}

func (f *Filesystem) vaultPath(vaultID string) (string, error) {
	if !vaultIDPattern.MatchString(vaultID) {
		return "", ErrUnsafePath
	}
	path := filepath.Join(f.root, vaultID)
	if !within(f.root, path) {
		return "", ErrUnsafePath
	}
	return path, nil
}
func (f *Filesystem) safePath(vaultID, relative string, allowMissing bool) (string, error) {
	if strings.Contains(strings.ToLower(relative), "%2e") || filepath.IsAbs(relative) || relative == "" {
		return "", ErrUnsafePath
	}
	root, err := f.vaultPath(vaultID)
	if err != nil {
		return "", err
	}
	clean := filepath.Clean(filepath.FromSlash(relative))
	if clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", ErrUnsafePath
	}
	path := filepath.Join(root, clean)
	if !within(root, path) {
		return "", ErrUnsafePath
	}
	if !allowMissing {
		resolved, e := filepath.EvalSymlinks(path)
		if e != nil {
			return "", e
		}
		if !within(root, resolved) {
			return "", ErrUnsafePath
		}
	}
	return path, nil
}
func (f *Filesystem) rejectSymlinkParents(vaultID, path string) error {
	root, err := f.vaultPath(vaultID)
	if err != nil {
		return err
	}
	current := filepath.Dir(path)
	for within(root, current) && current != root {
		info, e := os.Lstat(current)
		if e == nil && info.Mode()&os.ModeSymlink != 0 {
			return ErrUnsafePath
		}
		if e != nil && !os.IsNotExist(e) {
			return e
		}
		current = filepath.Dir(current)
	}
	resolved, e := filepath.EvalSymlinks(root)
	if e != nil {
		return fmt.Errorf("resolve vault: %w", e)
	}
	if !within(f.root, resolved) {
		return ErrUnsafePath
	}
	return nil
}
func within(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}
