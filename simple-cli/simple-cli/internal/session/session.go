// Package session provides the core Session entity, its status lifecycle,
// and the SessionStore abstraction for persistent session management.
// Constitution Principle I: independent package with clean public API.
package session

import (
	"encoding/json"
	"fmt"
	"time"
)

// SessionStatus is the lifecycle status of a Session.
type SessionStatus string

const (
	// StatusActive indicates the session is in active use.
	StatusActive SessionStatus = "active"
	// StatusPaused indicates the session exists but is not being actively used.
	StatusPaused SessionStatus = "paused"
	// StatusStopped indicates the session was explicitly stopped; state is retained.
	StatusStopped SessionStatus = "stopped"
)

// String returns the string representation of the status.
func (s SessionStatus) String() string { return string(s) }

// MarshalJSON implements json.Marshaler.
func (s SessionStatus) MarshalJSON() ([]byte, error) { return json.Marshal(string(s)) }

// UnmarshalJSON implements json.Unmarshaler.
func (s *SessionStatus) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("session: status: %w", err)
	}
	switch SessionStatus(raw) {
	case StatusActive, StatusPaused, StatusStopped:
		*s = SessionStatus(raw)
		return nil
	default:
		return fmt.Errorf("session: unknown status %q", raw)
	}
}

// Session represents a persistent user working context that survives terminal restarts.
// Fields use snake_case JSON tags matching the contract in contracts/output-schema.md.
type Session struct {
	// ID is a UUID v4 string; immutable after creation.
	ID string `json:"id"`
	// Name is a human-readable identifier; unique per user store.
	Name string `json:"name"`
	// Status is the lifecycle state.
	Status SessionStatus `json:"status"`
	// CreatedAt is the creation timestamp; immutable.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is updated on every mutation.
	UpdatedAt time.Time `json:"updated_at"`
	// State is a free-form key-value store; values must be JSON-serializable.
	State map[string]any `json:"state"`
}
