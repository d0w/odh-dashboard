package fleet

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

// Instance defines behavior required from every registry entry.
type Instance interface {
	http.Handler
	Close() error
}

// Factory creates an instance from its registry configuration.
type Factory[Config any, Entry Instance] func(context.Context, Config) (Entry, error)

// Registry dispatches requests to instances selected by a URL path key.
type Registry[Config any, Entry Instance] struct {
	factory Factory[Config, Entry]
	key     func(Config) string

	mu      sync.RWMutex
	entries map[string]Entry
}

func NewRegistry[Config any, Entry Instance](
	factory Factory[Config, Entry],
	key func(Config) string,
) *Registry[Config, Entry] {
	return &Registry[Config, Entry]{
		factory: factory,
		key:     key,
		entries: make(map[string]Entry),
	}
}

func (r *Registry[Config, Entry]) Register(ctx context.Context, config Config) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.key(config)
	if _, exists := r.entries[key]; exists {
		return false, nil
	}

	instance, err := r.factory(ctx, config)
	if err != nil {
		return false, fmt.Errorf("create registry entry: %w", err)
	}

	r.entries[key] = instance

	return true, nil
}

func (r *Registry[Config, Entry]) Get(key string) (Entry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instance, exists := r.entries[key]
	return instance, exists
}

func (r *Registry[Config, Entry]) GetEntries() []Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return flattenMap(r.entries)
}

func (r *Registry[Config, Entry]) Remove(_ context.Context, key string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.entries[key]; exists {
		delete(r.entries, key)
		return true, nil
	}

	return false, nil
}

func (r *Registry[Config, Entry]) Close(_ context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var closeErr error
	// TODO: Turn into goroutine WG
	for key, instance := range r.entries {
		if err := instance.Close(); err != nil {
			closeErr = errors.Join(closeErr, fmt.Errorf("close registry entry %v: %w", key, err))
		}
	}

	return closeErr
}

func (r *Registry[Config, Entry]) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	id, upstreamPath, upstreamRawPath, ok := splitPath(req.URL.Path, req.URL.EscapedPath())
	if !ok {
		return
	}

	instance, ok := r.Get(id)
	if !ok {
		_, _ = w.Write([]byte("Instance does not exist"))
		return
	}

	upstreamReq := req.Clone(req.Context())
	upstreamReq.URL.Path = upstreamPath
	upstreamReq.URL.RawPath = upstreamRawPath
	upstreamReq.RequestURI = upstreamReq.URL.RequestURI()
	instance.ServeHTTP(w, upstreamReq)
}
