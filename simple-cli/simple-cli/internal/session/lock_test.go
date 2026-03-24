package session_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/your-org/simple-cli/internal/session"
)

func TestFileLockAcquireRelease(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.lock")
	lock := session.NewFileLock(path)

	ctx := context.Background()
	require.NoError(t, lock.Lock(ctx))
	require.NoError(t, lock.Unlock())

	// Lock file should exist after lock/unlock.
	_, err := os.Stat(path)
	assert.NoError(t, err)
}

func TestFileLockContextTimeout(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.lock")

	// Acquire the lock with lock1.
	lock1 := session.NewFileLock(path)
	ctx := context.Background()
	require.NoError(t, lock1.Lock(ctx))
	defer lock1.Unlock() //nolint:errcheck

	// Attempt to acquire with lock2 using a short deadline.
	lock2 := session.NewFileLock(path)
	ctx2, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	err := lock2.Lock(ctx2)
	assert.ErrorIs(t, err, session.ErrLockTimeout)
}

func TestFileLockRelockAfterUnlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "relock.lock")
	lock := session.NewFileLock(path)
	ctx := context.Background()

	require.NoError(t, lock.Lock(ctx))
	require.NoError(t, lock.Unlock())

	// Should be able to lock again.
	lock2 := session.NewFileLock(path)
	require.NoError(t, lock2.Lock(ctx))
	require.NoError(t, lock2.Unlock())
}
