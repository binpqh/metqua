package tokenstore_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/binpqh/simple-cli/internal/provider"
	"github.com/binpqh/simple-cli/internal/tokenstore"
)

func TestFileTokenStore_SetGetDelete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")
	s := tokenstore.NewFileTokenStore(path)
	ctx := context.Background()

	ts := &provider.TokenSet{
		Provider:    "my-api",
		AccessToken: "at-123",
		Expiry:      time.Now().Add(1 * time.Hour),
		UserID:      "user-1",
	}

	if err := s.Set(ctx, "my-api", ts); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	got, err := s.Get(ctx, "my-api")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.AccessToken != ts.AccessToken || got.UserID != ts.UserID {
		t.Fatalf("mismatch: got %+v want %+v", got, ts)
	}

	if err := s.Delete(ctx, "my-api"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := s.Get(ctx, "my-api"); err == nil {
		t.Fatalf("expected ErrTokenNotFound after delete")
	}
}

func TestFileTokenStore_CorruptRecovery(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")
	// write corrupt file
	if err := os.WriteFile(path, []byte("notjson"), 0o600); err != nil {
		t.Fatalf("write corrupt: %v", err)
	}
	s := tokenstore.NewFileTokenStore(path)
	ctx := context.Background()
	// Get should return ErrTokenNotFound but not panic
	if _, err := s.Get(ctx, "any"); err == nil {
		t.Fatalf("expected error for missing token after corrupt recovery")
	}
}

func TestFileTokenStore_MultipleProviders(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")
	s := tokenstore.NewFileTokenStore(path)
	ctx := context.Background()

	a := &provider.TokenSet{Provider: "a", AccessToken: "at-a", Expiry: time.Now().Add(time.Hour)}
	b := &provider.TokenSet{Provider: "b", AccessToken: "at-b", Expiry: time.Now().Add(time.Hour)}

	if err := s.Set(ctx, "a", a); err != nil {
		t.Fatal(err)
	}
	if err := s.Set(ctx, "b", b); err != nil {
		t.Fatal(err)
	}

	ga, err := s.Get(ctx, "a")
	if err != nil || ga.AccessToken != "at-a" {
		t.Fatalf("expected at-a, got %+v, err %v", ga, err)
	}
	gb, err := s.Get(ctx, "b")
	if err != nil || gb.AccessToken != "at-b" {
		t.Fatalf("expected at-b, got %+v, err %v", gb, err)
	}
}

func TestFileTokenStore_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.json")
	s := tokenstore.NewFileTokenStore(path)
	ctx := context.Background()

	_, err := s.Get(ctx, "any")
	if err == nil {
		t.Fatalf("expected ErrTokenNotFound for missing file")
	}
}

func TestFileTokenStore_DeleteIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")
	s := tokenstore.NewFileTokenStore(path)
	ctx := context.Background()

	// Delete on empty store should not error
	if err := s.Delete(ctx, "nonexistent"); err != nil {
		t.Fatalf("Delete on empty store should not error: %v", err)
	}
}

func TestPathForConfigDir(t *testing.T) {
	got := tokenstore.PathForConfigDir("/some/dir")
	want := filepath.Join("/some/dir", "tokens.json")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
