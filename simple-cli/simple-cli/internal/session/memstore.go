package session

import (
	"context"
	"log/slog"
	"sync"
)

// MemStore is an in-memory SessionStore used as a fallback when the file system
// is unavailable. It emits a one-time warning on first use.
// It is safe for concurrent use via sync.RWMutex.
type MemStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session // id → Session
	names    map[string]string   // name → id
	warned   bool
}

// NewMemStore creates an empty MemStore.
func NewMemStore() *MemStore {
	return &MemStore{
		sessions: map[string]*Session{},
		names:    map[string]string{},
	}
}

func (m *MemStore) warnOnce() {
	if !m.warned {
		slog.Warn("session store is in-memory (read-only filesystem detected); data will not persist",
			"error", ErrStoreReadOnly)
		m.warned = true
	}
}

// Create stores a new session in memory.
func (m *MemStore) Create(_ context.Context, s *Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warnOnce()
	if _, exists := m.names[s.Name]; exists {
		return ErrNameConflict
	}
	clone := *s
	clone.State = copyState(s.State)
	m.sessions[s.ID] = &clone
	m.names[s.Name] = s.ID
	return nil
}

// Get retrieves a session by ID.
func (m *MemStore) Get(_ context.Context, id string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	if !ok {
		return nil, ErrNotFound
	}
	clone := *s
	clone.State = copyState(s.State)
	return &clone, nil
}

// GetByName retrieves a session by name.
func (m *MemStore) GetByName(ctx context.Context, name string) (*Session, error) {
	m.mu.RLock()
	id, ok := m.names[name]
	m.mu.RUnlock()
	if !ok {
		return nil, ErrNotFound
	}
	return m.Get(ctx, id)
}

// List returns all sessions, optionally filtered by status.
func (m *MemStore) List(_ context.Context, status *SessionStatus) ([]*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*Session
	for _, s := range m.sessions {
		if status != nil && s.Status != *status {
			continue
		}
		clone := *s
		clone.State = copyState(s.State)
		result = append(result, &clone)
	}
	return result, nil
}

// Update replaces an existing session.
func (m *MemStore) Update(_ context.Context, s *Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.sessions[s.ID]; !ok {
		return ErrNotFound
	}
	clone := *s
	clone.State = copyState(s.State)
	m.sessions[s.ID] = &clone
	m.names[s.Name] = s.ID
	return nil
}

// Delete removes a session from the store.
func (m *MemStore) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return ErrNotFound
	}
	delete(m.names, s.Name)
	delete(m.sessions, id)
	return nil
}

// Close is a no-op for MemStore.
func (m *MemStore) Close() error { return nil }

// copyState performs a shallow copy of the state map to prevent aliasing.
func copyState(src map[string]any) map[string]any {
	if src == nil {
		return map[string]any{}
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
