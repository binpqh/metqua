package session

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// indexFile maps session name → session ID for fast GetByName lookups.
type indexFile struct {
	Version int               `json:"version"`
	Entries map[string]string `json:"entries"` // name → id
}

// FileStore is a file-backed SessionStore implementation.
// Each session is persisted as <stateDir>/sessions/<uuid>.json.
// An index at <stateDir>/sessions/index.json maps names to IDs.
// Every write operation acquires the corresponding file lock first.
type FileStore struct {
	sessionsDir string
}

// NewFileStore initialises a FileStore rooted at stateDir.
// Returns (nil, ErrStoreReadOnly) when the directory cannot be created or written.
func NewFileStore(stateDir string) (*FileStore, error) {
	sessionsDir := filepath.Join(stateDir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0o700); err != nil {
		return nil, ErrStoreReadOnly
	}
	// Probe write access.
	probe := filepath.Join(sessionsDir, ".write-probe")
	if err := os.WriteFile(probe, []byte{}, 0o600); err != nil {
		return nil, ErrStoreReadOnly
	}
	_ = os.Remove(probe)
	return &FileStore{sessionsDir: sessionsDir}, nil
}

// sessionPath returns the JSON file path for a session ID.
func (fs *FileStore) sessionPath(id string) string {
	return filepath.Join(fs.sessionsDir, id+".json")
}

// lockPath returns the lock file path for a session ID.
func (fs *FileStore) lockPath(id string) string {
	return filepath.Join(fs.sessionsDir, id+".json.lock")
}

// indexPath returns the index file path.
func (fs *FileStore) indexPath() string { return filepath.Join(fs.sessionsDir, "index.json") }

// indexLockPath returns the lock file path for the index.
func (fs *FileStore) indexLockPath() string {
	return filepath.Join(fs.sessionsDir, "index.json.lock")
}

// ──────────────────────────────────────────────
// Index helpers
// ──────────────────────────────────────────────

func (fs *FileStore) readIndex() (*indexFile, error) {
	data, err := os.ReadFile(fs.indexPath())
	if os.IsNotExist(err) {
		return &indexFile{Version: 1, Entries: map[string]string{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read index: %w", err)
	}
	var idx indexFile
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parse index: %w", err)
	}
	if idx.Entries == nil {
		idx.Entries = map[string]string{}
	}
	return &idx, nil
}

func (fs *FileStore) writeIndex(idx *indexFile) error {
	b, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}
	return os.WriteFile(fs.indexPath(), b, 0o600)
}

// withIndexLock acquires the global index lock and calls fn.
func (fs *FileStore) withIndexLock(ctx context.Context, fn func(*indexFile) error) error {
	lk := NewFileLock(fs.indexLockPath())
	if err := lk.Lock(ctx); err != nil {
		return err
	}
	defer func() { _ = lk.Unlock() }()

	idx, err := fs.readIndex()
	if err != nil {
		return err
	}
	return fn(idx)
}

// ──────────────────────────────────────────────
// SessionStore implementation
// ──────────────────────────────────────────────

// Create persists a new session. ErrNameConflict if the name is taken.
func (fs *FileStore) Create(ctx context.Context, s *Session) error {
	return fs.withIndexLock(ctx, func(idx *indexFile) error {
		if _, exists := idx.Entries[s.Name]; exists {
			return ErrNameConflict
		}
		if err := fs.writeSession(ctx, s); err != nil {
			return err
		}
		idx.Entries[s.Name] = s.ID
		return fs.writeIndex(idx)
	})
}

// Get retrieves a session by ID.
func (fs *FileStore) Get(_ context.Context, id string) (*Session, error) {
	data, err := os.ReadFile(fs.sessionPath(id))
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read session %s: %w", id, err)
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse session %s: %w", id, err)
	}
	return &s, nil
}

// GetByName retrieves a session by name.
func (fs *FileStore) GetByName(ctx context.Context, name string) (*Session, error) {
	idx, err := fs.readIndex()
	if err != nil {
		return nil, err
	}
	id, ok := idx.Entries[name]
	if !ok {
		return nil, ErrNotFound
	}
	return fs.Get(ctx, id)
}

// List returns all sessions, optionally filtered by status.
func (fs *FileStore) List(_ context.Context, status *SessionStatus) ([]*Session, error) {
	entries, err := os.ReadDir(fs.sessionsDir)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	var result []*Session
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" || e.Name() == "index.json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(fs.sessionsDir, e.Name()))
		if err != nil {
			slog.Warn("skipping unreadable session file", "file", e.Name(), "err", err)
			continue
		}
		var s Session
		if err := json.Unmarshal(data, &s); err != nil {
			slog.Warn("skipping unparseable session file", "file", e.Name(), "err", err)
			continue
		}
		if status != nil && s.Status != *status {
			continue
		}
		result = append(result, &s)
	}
	return result, nil
}

// Update replaces an existing session record.
func (fs *FileStore) Update(ctx context.Context, s *Session) error {
	if _, err := os.Stat(fs.sessionPath(s.ID)); os.IsNotExist(err) {
		return ErrNotFound
	}
	return fs.writeSession(ctx, s)
}

// Delete removes a session file and its index entry.
func (fs *FileStore) Delete(ctx context.Context, id string) error {
	return fs.withIndexLock(ctx, func(idx *indexFile) error {
		lk := NewFileLock(fs.lockPath(id))
		if err := lk.Lock(ctx); err != nil {
			return err
		}
		defer func() { _ = lk.Unlock() }()

		if err := os.Remove(fs.sessionPath(id)); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("delete session %s: %w", id, err)
		}
		_ = os.Remove(fs.lockPath(id))

		// Remove from index.
		for name, sid := range idx.Entries {
			if sid == id {
				delete(idx.Entries, name)
				break
			}
		}
		return fs.writeIndex(idx)
	})
}

// Close is a no-op for FileStore — file handles are opened and closed per operation.
func (fs *FileStore) Close() error { return nil }

// writeSession atomically writes a session JSON file under its per-session lock.
func (fs *FileStore) writeSession(ctx context.Context, s *Session) error {
	lk := NewFileLock(fs.lockPath(s.ID))
	if err := lk.Lock(ctx); err != nil {
		return err
	}
	defer func() { _ = lk.Unlock() }()

	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session %s: %w", s.ID, err)
	}
	// Write to temp file then rename for atomicity.
	tmp := fs.sessionPath(s.ID) + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return fmt.Errorf("write session %s: %w", s.ID, err)
	}
	return os.Rename(tmp, fs.sessionPath(s.ID))
}

// lastWrite tracks whether ErrStoreReadOnly warning was already emitted.
var _ = time.Now // silence unused import on older linters
