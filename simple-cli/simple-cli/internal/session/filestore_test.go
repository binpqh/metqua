package session_test

import (
	"context"
	"os"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/your-org/simple-cli/internal/session"
)

func newFileStore(t *testing.T) *session.FileStore {
	t.Helper()
	dir := t.TempDir()
	fs, err := session.NewFileStore(dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = fs.Close() })
	return fs
}

func TestFileStoreCreate(t *testing.T) {
	ctx := context.Background()
	store := newFileStore(t)

	s := newTestSession("file-alpha")
	require.NoError(t, store.Create(ctx, s))

	// Duplicate name.
	err := store.Create(ctx, s)
	assert.ErrorIs(t, err, session.ErrNameConflict)
}

func TestFileStoreGet(t *testing.T) {
	ctx := context.Background()
	store := newFileStore(t)

	s := newTestSession("file-beta")
	require.NoError(t, store.Create(ctx, s))

	got, err := store.Get(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, s.ID, got.ID)
	assert.Equal(t, s.Name, got.Name)
}

func TestFileStoreGetNotFound(t *testing.T) {
	ctx := context.Background()
	store := newFileStore(t)
	_, err := store.Get(ctx, "no-such-id")
	assert.ErrorIs(t, err, session.ErrNotFound)
}

func TestFileStoreGetByName(t *testing.T) {
	ctx := context.Background()
	store := newFileStore(t)

	s := newTestSession("file-gamma")
	require.NoError(t, store.Create(ctx, s))

	got, err := store.GetByName(ctx, "file-gamma")
	require.NoError(t, err)
	assert.Equal(t, s.ID, got.ID)
}

func TestFileStoreList(t *testing.T) {
	ctx := context.Background()
	store := newFileStore(t)

	require.NoError(t, store.Create(ctx, newTestSession("file-one")))
	require.NoError(t, store.Create(ctx, newTestSession("file-two")))

	s3 := newTestSession("file-three")
	s3.Status = session.StatusStopped
	require.NoError(t, store.Create(ctx, s3))

	all, err := store.List(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, all, 3)

	active := session.StatusActive
	filtered, err := store.List(ctx, &active)
	require.NoError(t, err)
	assert.Len(t, filtered, 2)
}

func TestFileStoreUpdate(t *testing.T) {
	ctx := context.Background()
	store := newFileStore(t)

	s := newTestSession("file-delta")
	require.NoError(t, store.Create(ctx, s))

	s.Status = session.StatusStopped
	require.NoError(t, store.Update(ctx, s))

	got, err := store.Get(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, session.StatusStopped, got.Status)
}

func TestFileStoreDelete(t *testing.T) {
	ctx := context.Background()
	store := newFileStore(t)

	s := newTestSession("file-epsilon")
	require.NoError(t, store.Create(ctx, s))
	require.NoError(t, store.Delete(ctx, s.ID))

	_, err := store.Get(ctx, s.ID)
	assert.ErrorIs(t, err, session.ErrNotFound)
}

func TestFileStorePersistence(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	// Create and close store.
	store1, err := session.NewFileStore(dir)
	require.NoError(t, err)
	s := newTestSession("persist-me")
	require.NoError(t, store1.Create(ctx, s))
	require.NoError(t, store1.Close())

	// Reopen store.
	store2, err := session.NewFileStore(dir)
	require.NoError(t, err)
	defer store2.Close() //nolint:errcheck

	got, err := store2.GetByName(ctx, "persist-me")
	require.NoError(t, err)
	assert.Equal(t, s.ID, got.ID)
}

func TestFileStoreReadOnly(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root — permissions not enforced")
	}
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not enforce chmod directory permissions for directory creation")
	}
	// Create a read-only directory.
	dir := t.TempDir()
	require.NoError(t, os.Chmod(dir, 0o444))
	t.Cleanup(func() { os.Chmod(dir, 0o755) }) //nolint:errcheck

	_, err := session.NewFileStore(dir)
	assert.Error(t, err)
}

// TestFileStoreConcurrentWrites ensures concurrent creates are safe.
func TestFileStoreConcurrentWrites(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newFileStore(t)

	const n = 20
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			s := newTestSession("concurrent-" + itoa(i))
			_ = store.Create(ctx, s)
		}(i)
	}
	wg.Wait()

	all, err := store.List(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, all, n)
}
