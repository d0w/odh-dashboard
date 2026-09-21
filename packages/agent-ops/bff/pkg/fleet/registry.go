package fleet

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
)

type GatewayConfig struct {
	ID       string
	Endpoint string
}

type GatewayInstance struct {
	handler http.Handler
	Close   func() error
}

// TODO: make package-agnostic. no gateway specific stuff just registry
// eventually change to interface to allow any object to be created
type GatewayFactory func(context.Context, GatewayConfig) (GatewayInstance, error)

type Registry interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
	Register(context.Context, GatewayConfig) (bool, error)
	Remove(context.Context, string) (bool, error)
	Close(context.Context) error
}

type GatewayRegistry struct {
	factory GatewayFactory

	mu      sync.Mutex
	entries map[string]*GatewayInstance
}

func NewRegistry(factory GatewayFactory) *GatewayRegistry {
	return &GatewayRegistry{
		factory: factory,
		entries: make(map[string]*GatewayInstance),
	}
}

func (r *GatewayRegistry) Register(ctx context.Context, config GatewayConfig) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.entries[config.ID]
	if exists {
		return false, nil
	}

	instance, err := r.factory(ctx, config)
	if err != nil {
		return false, fmt.Errorf("create gateway %q: %w", config.ID, err)
	}

	r.entries[config.ID] = &instance

	return true, nil
}

func (r *GatewayRegistry) Remove(ctx context.Context, id string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.entries[id]
	if exists {
		delete(r.entries, id)
		return true, nil
	}

	return false, nil
}

func (r *GatewayRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	id, upstreamPath, upstreamRawPath, ok := splitGatewayPath(req.URL.Path, req.URL.EscapedPath())
	slog.Info("HIT THE REGISTRY")
	if !ok {
		// return http error
		return
	}
	// if !validGatewayID(id) {
	// 	// return http error
	// 	return
	// }

	gatewayInstance, ok := r.entries[id]
	if !ok {
		w.Write([]byte("Instance does not exist"))
		return
	}

	// clones request
	// alter headers, auth, etc. here...
	upstreamReq := req.Clone(req.Context())
	upstreamReq.URL.Path = upstreamPath
	upstreamReq.URL.RawPath = upstreamRawPath
	upstreamReq.RequestURI = upstreamReq.URL.RequestURI()
	gatewayInstance.handler.ServeHTTP(w, upstreamReq)
}

func (r *GatewayRegistry) Close(ctx context.Context) error {
	// close all gateway instances. be aware of race conditions
	return nil
	// errChan := make(chan error, 1)
	// select {
	// case <-ctx.Done():
	// 	return nil
	// case err := <-errChan:
	// 	return err
	// }
}
