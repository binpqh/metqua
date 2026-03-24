package session

import "context"

// SessionStore defines the persistence contract for sessions.
// All implementations MUST be safe for concurrent use within a single process
// (cross-process safety is provided by OS-level file locking in FileStore).
type SessionStore interface {
	// Create persists a new session. Returns ErrNameConflict if the name is taken.
	Create(ctx context.Context, s *Session) error

	// Get retrieves a session by ID. Returns ErrNotFound if it does not exist.
	Get(ctx context.Context, id string) (*Session, error)

	// GetByName retrieves a session by name. Returns ErrNotFound if it does not exist.
	GetByName(ctx context.Context, name string) (*Session, error)

	// List returns all sessions. When status is non-nil, results are filtered.
	List(ctx context.Context, status *SessionStatus) ([]*Session, error)

	// Update replaces the full session record. Returns ErrNotFound if absent.
	Update(ctx context.Context, s *Session) error

	// Delete permanently removes a session and its state. Returns ErrNotFound if absent.
	Delete(ctx context.Context, id string) error

	// Close releases any held resources (file locks, file handles).
	Close() error
}
