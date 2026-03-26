package tokenstore

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/binpqh/simple-cli/internal/provider"
)

// FileTokenStore persists tokens in a single JSON file at Path.
type FileTokenStore struct {
	Path string
}

// fileLayout mirrors on-disk structure.
type fileLayout struct {
	Providers map[string]*provider.TokenSet `json:"providers"`
}

// NewFileTokenStore returns a FileTokenStore using the given path.
func NewFileTokenStore(path string) *FileTokenStore {
	return &FileTokenStore{Path: path}
}

// read loads the file layout, returns empty layout if file not found.
func (f *FileTokenStore) read() (*fileLayout, error) {
	data, err := os.ReadFile(f.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return &fileLayout{Providers: map[string]*provider.TokenSet{}}, nil
		}
		return nil, err
	}
	var l fileLayout
	if err := json.Unmarshal(data, &l); err != nil {
		// treat corrupt file as not found so caller can recover
		_ = os.Remove(f.Path)
		return &fileLayout{Providers: map[string]*provider.TokenSet{}}, nil
	}
	if l.Providers == nil {
		l.Providers = map[string]*provider.TokenSet{}
	}
	return &l, nil
}

// write persists layout atomically.
func (f *FileTokenStore) write(l *fileLayout) error {
	dir := filepath.Dir(f.Path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	// write temp file then rename
	tmp := f.Path + ".tmp"
	if err := os.WriteFile(tmp, data, fs.FileMode(0o600)); err != nil {
		return err
	}
	return os.Rename(tmp, f.Path)
}

func (f *FileTokenStore) Get(ctx context.Context, providerName string) (*provider.TokenSet, error) {
	l, err := f.read()
	if err != nil {
		return nil, err
	}
	t, ok := l.Providers[providerName]
	if !ok || t == nil {
		return nil, provider.ErrTokenNotFound
	}
	return t, nil
}

func (f *FileTokenStore) Set(ctx context.Context, providerName string, t *provider.TokenSet) error {
	l, err := f.read()
	if err != nil {
		return err
	}
	l.Providers[providerName] = t
	return f.write(l)
}

func (f *FileTokenStore) Delete(ctx context.Context, providerName string) error {
	l, err := f.read()
	if err != nil {
		return err
	}
	delete(l.Providers, providerName)
	return f.write(l)
}

// PathForConfigDir returns default path for tokens.json under configDir.
func PathForConfigDir(configDir string) string {
	return filepath.Join(configDir, "tokens.json")
}

// Ensure FileTokenStore implements provider.TokenStore
var _ provider.TokenStore = (*FileTokenStore)(nil)
