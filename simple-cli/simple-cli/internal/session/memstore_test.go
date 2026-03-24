package session_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/your-org/simple-cli/internal/session"
)

func newTestSession(name string) *session.Session {
	return &session.Session{
		ID:        name + "-id",
		Name:      name,
		Status:    session.StatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		State:     map[string]any{},
	}
}

func TestMemStoreCreate(t *testing.T) {
	ctx := context.Background()
	store := session.NewMemStore()

	s := newTestSession("alpha")
	require.NoError(t, store.Create(ctx, s))

	// Duplicate name returns ErrNameConflict.
	err := store.Create(ctx, s)
	assert.ErrorIs(t, err, session.ErrNameConflict)
}

func TestMemStoreGet(t *testing.T) {
	ctx := context.Background()
	store := session.NewMemStore()

	s := newTestSession("beta")
	require.NoError(t, store.Create(ctx, s))

	got, err := store.Get(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, s.ID, got.ID)
	assert.Equal(t, s.Name, got.Name)
}

func TestMemStoreGetNotFound(t *testing.T) {
	ctx := context.Background()
	store := session.NewMemStore()
	_, err := store.Get(ctx, "no-such-id")
	assert.ErrorIs(t, err, session.ErrNotFound)
}

func TestMemStoreGetByName(t *testing.T) {
	ctx := context.Background()
	store := session.NewMemStore()

	s := newTestSession("gamma")
	require.NoError(t, store.Create(ctx, s))

	got, err := store.GetByName(ctx, "gamma")
	require.NoError(t, err)
	assert.Equal(t, s.ID, got.ID)
}

func TestMemStoreList(t *testing.T) {
	ctx := context.Background()
	store := session.NewMemStore()

	require.NoError(t, store.Create(ctx, newTestSession("one")))
	require.NoError(t, store.Create(ctx, newTestSession("two")))

	s3 := newTestSession("three")
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

func TestMemStoreUpdate(t *testing.T) {
	ctx := context.Background()
	store := session.NewMemStore()

	s := newTestSession("delta")
	require.NoError(t, store.Create(ctx, s))

	s.Status = session.StatusStopped
	require.NoError(t, store.Update(ctx, s))

	got, err := store.Get(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, session.StatusStopped, got.Status)
}

func TestMemStoreUpdateNotFound(t *testing.T) {
	ctx := context.Background()
	store := session.NewMemStore()
	s := newTestSession("ghost")
	err := store.Update(ctx, s)
	assert.ErrorIs(t, err, session.ErrNotFound)
}

func TestMemStoreDelete(t *testing.T) {
	ctx := context.Background()
	store := session.NewMemStore()

	s := newTestSession("epsilon")
	require.NoError(t, store.Create(ctx, s))
	require.NoError(t, store.Delete(ctx, s.ID))

	_, err := store.Get(ctx, s.ID)
	assert.ErrorIs(t, err, session.ErrNotFound)
}

func TestMemStoreDeleteNotFound(t *testing.T) {
	ctx := context.Background()
	store := session.NewMemStore()
	err := store.Delete(ctx, "no-such-id")
	assert.ErrorIs(t, err, session.ErrNotFound)
}

// TestMemStoreConcurrency verifies concurrent reads and writes are safe.
func TestMemStoreClose(t *testing.T) {
	store := session.NewMemStore()
	assert.NoError(t, store.Close())
}

// TestProxyStore verifies ProxyStore delegates to the underlying store.
func TestProxyStore(t *testing.T) {
	ctx := context.Background()
	mem := session.NewMemStore()
	var iface session.SessionStore = mem

	proxy := session.NewProxyStore(&iface)

	s := newTestSession("proxy-test")
	require.NoError(t, proxy.Create(ctx, s))

	got, err := proxy.Get(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, s.ID, got.ID)

	got2, err := proxy.GetByName(ctx, s.Name)
	require.NoError(t, err)
	assert.Equal(t, s.ID, got2.ID)

	all, err := proxy.List(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	s.Status = session.StatusStopped
	require.NoError(t, proxy.Update(ctx, s))

	require.NoError(t, proxy.Delete(ctx, s.ID))
	require.NoError(t, proxy.Close())
}

// TestProxyStoreNilDelegate verifies ProxyStore fails gracefully when uninitialised.
func TestProxyStoreNilDelegate(t *testing.T) {
	ctx := context.Background()
	var iface session.SessionStore // nil
	proxy := session.NewProxyStore(&iface)

	_, err := proxy.Get(ctx, "x")
	assert.Error(t, err)
}

// TestMemStoreConcurrency verifies concurrent reads and writes are safe.
func TestMemStoreConcurrency(t *testing.T) {
	ctx := context.Background()
	store := session.NewMemStore()

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			name := "sess-" + itoa(i)
			s := newTestSession(name)
			if err := store.Create(ctx, s); err != nil {
				return
			}
			_, _ = store.Get(ctx, s.ID)
			_, _ = store.GetByName(ctx, name)
			_, _ = store.List(ctx, nil)
		}(i)
	}
	wg.Wait()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	b := make([]byte, 0, 4)
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
