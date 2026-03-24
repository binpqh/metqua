package session

import (
	"context"
	"fmt"
)

// ProxyStore is a SessionStore that delegates to the store pointed at by *delegate.
// This lets cmd/root.go register the cobra command in init() (before the
// real store is initialised) and swap in the real store during
// PersistentPreRunE.
type ProxyStore struct {
	delegate *SessionStore
}

// NewProxyStore wraps a pointer to a SessionStore interface value. The real
// store must be assigned to *delegate before any method is called.
func NewProxyStore(delegate *SessionStore) *ProxyStore {
	return &ProxyStore{delegate: delegate}
}

func (p *ProxyStore) real() (SessionStore, error) {
	if p.delegate == nil || *p.delegate == nil {
		return nil, fmt.Errorf("session store not yet initialised")
	}
	return *p.delegate, nil
}

func (p *ProxyStore) Create(ctx context.Context, s *Session) error {
	r, err := p.real()
	if err != nil {
		return err
	}
	return r.Create(ctx, s)
}

func (p *ProxyStore) Get(ctx context.Context, id string) (*Session, error) {
	r, err := p.real()
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

func (p *ProxyStore) GetByName(ctx context.Context, name string) (*Session, error) {
	r, err := p.real()
	if err != nil {
		return nil, err
	}
	return r.GetByName(ctx, name)
}

func (p *ProxyStore) List(ctx context.Context, status *SessionStatus) ([]*Session, error) {
	r, err := p.real()
	if err != nil {
		return nil, err
	}
	return r.List(ctx, status)
}

func (p *ProxyStore) Update(ctx context.Context, s *Session) error {
	r, err := p.real()
	if err != nil {
		return err
	}
	return r.Update(ctx, s)
}

func (p *ProxyStore) Delete(ctx context.Context, id string) error {
	r, err := p.real()
	if err != nil {
		return err
	}
	return r.Delete(ctx, id)
}

func (p *ProxyStore) Close() error {
	if p.delegate == nil || *p.delegate == nil {
		return nil
	}
	return (*p.delegate).Close()
}
