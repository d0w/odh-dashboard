package tempbff

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

const openshellPrefix = "/api/openshell"

var gatewayIDPattern = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

var ErrInvalidGatewayID = errors.New("gateway ID must be an RFC 1123 label")

// GatewayConfig identifies a gateway registered by a trusted control plane.
// Endpoint validation belongs in the production factory because allowed targets
// depend on deployment-specific service and TLS policy.
type GatewayConfig struct {
	ID       string
	Endpoint string
}

// GatewayInstance wraps one upstream server.App and resources it owns, such as
// OpenShell SDK and raw-exec gRPC clients.
type GatewayInstance struct {
	Handler http.Handler
	Close   func() error
}

// GatewayFactory creates an isolated upstream server for one gateway.
type GatewayFactory func(context.Context, GatewayConfig) (GatewayInstance, error)

// GatewayRegistrar is the contract used by future gateway discovery and
// reconciliation services. Registration remains internal; HTTP clients never
// receive this capability.
type GatewayRegistrar interface {
	Register(context.Context, GatewayConfig) (bool, error)
	Remove(context.Context, string) (bool, error)
}

// GatewayRouter is the HTTP dispatch contract mounted by the BFF's root mux.
type GatewayRouter interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// GatewayRegistry combines lifecycle, registration, and request dispatch.
// The root BFF owns Close; discovery services should depend only on
// GatewayRegistrar.
type GatewayRegistry interface {
	GatewayRegistrar
	GatewayRouter
	Close(context.Context) error
}

// Registry dynamically dispatches gateway routes without mutating an HTTP
// router after startup. Register and Remove are intended for a trusted
// discovery/reconciliation service, not browser-facing handlers.
type Registry struct {
	factory GatewayFactory

	mu         sync.Mutex
	operations sync.Mutex
	entries    map[string]*gatewayEntry
}

type gatewayEntry struct {
	handler http.Handler
	close   func() error

	refs      int
	retiring  bool
	drained   chan struct{}
	closeOnce sync.Once
	closeErr  error
}

var _ GatewayRegistry = (*Registry)(nil)

// NewRegistry constructs an empty registry.
func NewRegistry(factory GatewayFactory) *Registry {
	return &Registry{
		factory: factory,
		entries: make(map[string]*gatewayEntry),
	}
}

// Register creates a gateway only when its ID is absent. Existing IDs are
// deliberately left unchanged; config rotation needs an explicit replacement
// operation rather than silently changing a request target.
func (r *Registry) Register(ctx context.Context, config GatewayConfig) (bool, error) {
	if err := validateGatewayConfig(config); err != nil {
		return false, err
	}
	if r.factory == nil {
		return false, errors.New("gateway factory is required")
	}

	r.operations.Lock()
	defer r.operations.Unlock()

	r.mu.Lock()
	_, exists := r.entries[config.ID]
	r.mu.Unlock()
	if exists {
		return false, nil
	}

	instance, err := r.factory(ctx, config)
	if err != nil {
		return false, fmt.Errorf("create gateway %q: %w", config.ID, err)
	}
	if instance.Handler == nil {
		if instance.Close != nil {
			_ = instance.Close()
		}
		return false, fmt.Errorf("create gateway %q: handler is required", config.ID)
	}
	if instance.Close == nil {
		instance.Close = func() error { return nil }
	}

	r.mu.Lock()
	r.entries[config.ID] = &gatewayEntry{
		handler: instance.Handler,
		close:   instance.Close,
		drained: make(chan struct{}),
	}
	r.mu.Unlock()

	return true, nil
}

// Remove unpublishes a gateway before waiting for active requests and closing
// its clients. A canceled context still schedules eventual cleanup.
func (r *Registry) Remove(ctx context.Context, id string) (bool, error) {
	if !validGatewayID(id) {
		return false, ErrInvalidGatewayID
	}

	r.operations.Lock()
	defer r.operations.Unlock()

	r.mu.Lock()
	entry, exists := r.entries[id]
	if exists {
		delete(r.entries, id)
		r.retireLocked(entry)
	}
	r.mu.Unlock()
	if !exists {
		return false, nil
	}

	return true, r.waitAndClose(ctx, entry)
}

// Close unpublishes every gateway and drains their active requests.
func (r *Registry) Close(ctx context.Context) error {
	r.operations.Lock()
	defer r.operations.Unlock()

	r.mu.Lock()
	entries := make([]*gatewayEntry, 0, len(r.entries))
	for id, entry := range r.entries {
		delete(r.entries, id)
		r.retireLocked(entry)
		entries = append(entries, entry)
	}
	r.mu.Unlock()

	var closeErr error
	for _, entry := range entries {
		closeErr = errors.Join(closeErr, r.waitAndClose(ctx, entry))
	}
	return closeErr
}

// ServeHTTP expects a path after /api/openshell, for example
// /gateway-a/api/v1/workspaces. It rewrites only the path passed to the
// upstream handler; headers, request context, body, and query stay intact.
func (r *Registry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	id, upstreamPath, upstreamRawPath, ok := splitGatewayPath(req.URL.Path, req.URL.EscapedPath())
	if !ok {
		writeError(w, http.StatusBadRequest, "gateway ID is required")
		return
	}
	if !validGatewayID(id) {
		writeError(w, http.StatusBadRequest, ErrInvalidGatewayID.Error())
		return
	}

	entry := r.acquire(id)
	if entry == nil {
		writeError(w, http.StatusNotFound, "gateway not found")
		return
	}
	defer r.release(entry)

	upstreamReq := req.Clone(req.Context())
	upstreamReq.URL.Path = upstreamPath
	upstreamReq.URL.RawPath = upstreamRawPath
	upstreamReq.RequestURI = upstreamReq.URL.RequestURI()
	entry.handler.ServeHTTP(w, upstreamReq)
}

// NewRouter mounts registry dispatch below the public multi-gateway prefix.
func NewRouter(registry *Registry) http.Handler {
	router := http.NewServeMux()
	router.Handle(openshellPrefix+"/", http.StripPrefix(openshellPrefix, registry))
	return router
}

func validateGatewayConfig(config GatewayConfig) error {
	if !validGatewayID(config.ID) {
		return ErrInvalidGatewayID
	}
	if config.Endpoint == "" {
		return errors.New("gateway endpoint is required")
	}
	return nil
}

func validGatewayID(id string) bool {
	return len(id) <= 63 && gatewayIDPattern.MatchString(id)
}

func (r *Registry) acquire(id string) *gatewayEntry {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry := r.entries[id]
	if entry == nil || entry.retiring {
		return nil
	}
	entry.refs++
	return entry
}

func (r *Registry) release(entry *gatewayEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry.refs--
	if entry.retiring && entry.refs == 0 {
		close(entry.drained)
	}
}

func (r *Registry) retireLocked(entry *gatewayEntry) {
	entry.retiring = true
	if entry.refs == 0 {
		close(entry.drained)
	}
}

func (r *Registry) waitAndClose(ctx context.Context, entry *gatewayEntry) error {
	select {
	case <-entry.drained:
		return entry.closeResources()
	case <-ctx.Done():
		go func() {
			<-entry.drained
			_ = entry.closeResources()
		}()
		return ctx.Err()
	}
}

func (entry *gatewayEntry) closeResources() error {
	entry.closeOnce.Do(func() {
		entry.closeErr = entry.close()
	})
	return entry.closeErr
}

func splitGatewayPath(requestPath, escapedPath string) (string, string, string, bool) {
	trimmedPath := strings.TrimPrefix(requestPath, "/")
	id, remainingPath, hasRemainingPath := strings.Cut(trimmedPath, "/")
	if id == "" {
		return "", "", "", false
	}
	if hasRemainingPath {
		remainingPath = "/" + remainingPath
	} else {
		remainingPath = "/"
	}

	trimmedEscapedPath := strings.TrimPrefix(escapedPath, "/")
	_, remainingEscapedPath, hasRemainingEscapedPath := strings.Cut(trimmedEscapedPath, "/")
	if hasRemainingEscapedPath {
		remainingEscapedPath = "/" + remainingEscapedPath
	} else {
		remainingEscapedPath = ""
	}

	return id, remainingPath, remainingEscapedPath, true
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
