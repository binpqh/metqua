package session

import "errors"

// Sentinel errors returned by SessionStore implementations.
// Callers MUST use errors.Is for comparison.
var (
	// ErrNotFound is returned when the requested session does not exist.
	ErrNotFound = errors.New("session not found")

	// ErrNameConflict is returned when a session with the same name already exists.
	ErrNameConflict = errors.New("session name already in use")

	// ErrLockTimeout is returned when a file lock cannot be acquired within the deadline.
	ErrLockTimeout = errors.New("could not acquire session lock within timeout")

	// ErrStoreReadOnly is returned when the state directory is not writable.
	// The store falls back to in-memory operation.
	ErrStoreReadOnly = errors.New("session store is read-only; operating in-memory")

	// ErrSessionStopped is returned when a stopped session is resumed directly.
	ErrSessionStopped = errors.New("session is stopped; use 'session reset' to start fresh")
)
