package session_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/your-org/simple-cli/internal/session"
)

func TestSessionStatusString(t *testing.T) {
	assert.Equal(t, "active", session.StatusActive.String())
	assert.Equal(t, "paused", session.StatusPaused.String())
	assert.Equal(t, "stopped", session.StatusStopped.String())
}

func TestSessionStatusJSONRoundTrip(t *testing.T) {
	for _, status := range []session.SessionStatus{
		session.StatusActive,
		session.StatusPaused,
		session.StatusStopped,
	} {
		t.Run(string(status), func(t *testing.T) {
			b, err := json.Marshal(status)
			require.NoError(t, err)

			var got session.SessionStatus
			require.NoError(t, json.Unmarshal(b, &got))
			assert.Equal(t, status, got)
		})
	}
}

func TestSessionStatusUnmarshalUnknown(t *testing.T) {
	var s session.SessionStatus
	err := json.Unmarshal([]byte(`"unknown-status"`), &s)
	assert.Error(t, err)
}

func TestSessionJSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	original := &session.Session{
		ID:        "abc-123",
		Name:      "my-session",
		Status:    session.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
		State:     map[string]any{"key": "value", "count": float64(42)},
	}

	b, err := json.Marshal(original)
	require.NoError(t, err)

	var got session.Session
	require.NoError(t, json.Unmarshal(b, &got))

	assert.Equal(t, original.ID, got.ID)
	assert.Equal(t, original.Name, got.Name)
	assert.Equal(t, original.Status, got.Status)
	assert.Equal(t, original.State["key"], got.State["key"])
	assert.Equal(t, original.State["count"], got.State["count"])
}
