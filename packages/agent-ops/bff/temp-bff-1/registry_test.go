package tempbff

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRegistryRoutesGatewayAndPreservesRequest(t *testing.T) {
	var factoryCalls atomic.Int32
	var gotPath string
	var gotQuery string

	registry := NewRegistry(func(_ context.Context, config GatewayConfig) (GatewayInstance, error) {
		factoryCalls.Add(1)
		return GatewayInstance{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				gotPath = req.URL.Path
				gotQuery = req.URL.RawQuery
				w.Header().Set("X-Gateway-ID", config.ID)
				w.WriteHeader(http.StatusNoContent)
			}),
		}, nil
	})

	created, err := registry.Register(context.Background(), GatewayConfig{ID: "gateway-a", Endpoint: "grpcs://gateway-a:50051"})
	if err != nil || !created {
		t.Fatalf("Register() = (%t, %v), want (true, nil)", created, err)
	}
	created, err = registry.Register(context.Background(), GatewayConfig{ID: "gateway-a", Endpoint: "grpcs://other:50051"})
	if err != nil || created {
		t.Fatalf("duplicate Register() = (%t, %v), want (false, nil)", created, err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/openshell/gateway-a/api/v1/gateway?include=drivers", nil)
	NewRouter(registry).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if recorder.Header().Get("X-Gateway-ID") != "gateway-a" {
		t.Fatalf("X-Gateway-ID = %q, want gateway-a", recorder.Header().Get("X-Gateway-ID"))
	}
	if gotPath != "/api/v1/gateway" {
		t.Fatalf("upstream path = %q, want /api/v1/gateway", gotPath)
	}
	if gotQuery != "include=drivers" {
		t.Fatalf("upstream query = %q, want include=drivers", gotQuery)
	}
	if factoryCalls.Load() != 1 {
		t.Fatalf("factory calls = %d, want 1", factoryCalls.Load())
	}
}

func TestRegistryRejectsUnknownAndMalformedGatewayIDs(t *testing.T) {
	registry := NewRegistry(func(context.Context, GatewayConfig) (GatewayInstance, error) {
		return GatewayInstance{Handler: http.NotFoundHandler()}, nil
	})
	router := NewRouter(registry)

	for _, test := range []struct {
		name string
		path string
		want int
	}{
		{name: "unknown", path: "/api/openshell/missing/api/v1/gateway", want: http.StatusNotFound},
		{name: "missing", path: "/api/openshell/", want: http.StatusBadRequest},
		{name: "malformed", path: "/api/openshell/Bad_ID/api/v1/gateway", want: http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
			if recorder.Code != test.want {
				t.Fatalf("status = %d, want %d", recorder.Code, test.want)
			}
		})
	}
}

func TestRegistryRegistersConcurrentDuplicateOnce(t *testing.T) {
	started := make(chan struct{})
	allowCreation := make(chan struct{})
	var factoryCalls atomic.Int32

	registry := NewRegistry(func(context.Context, GatewayConfig) (GatewayInstance, error) {
		factoryCalls.Add(1)
		close(started)
		<-allowCreation
		return GatewayInstance{Handler: http.NotFoundHandler()}, nil
	})

	results := make(chan bool, 2)
	registration := func() {
		created, err := registry.Register(context.Background(), GatewayConfig{ID: "gateway-a", Endpoint: "grpcs://gateway-a:50051"})
		if err != nil {
			t.Errorf("Register() error = %v", err)
		}
		results <- created
	}
	go registration()
	<-started
	go registration()
	close(allowCreation)

	first, second := <-results, <-results
	if first == second {
		t.Fatalf("concurrent Register() results = (%t, %t), want one creation", first, second)
	}
	if factoryCalls.Load() != 1 {
		t.Fatalf("factory calls = %d, want 1", factoryCalls.Load())
	}
}

func TestRemoveDrainsRequestBeforeClosingGateway(t *testing.T) {
	started := make(chan struct{})
	allowResponse := make(chan struct{})
	closed := make(chan struct{})

	registry := NewRegistry(func(context.Context, GatewayConfig) (GatewayInstance, error) {
		return GatewayInstance{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				close(started)
				<-allowResponse
				w.WriteHeader(http.StatusNoContent)
			}),
			Close: func() error {
				close(closed)
				return nil
			},
		}, nil
	})
	_, err := registry.Register(context.Background(), GatewayConfig{ID: "gateway-a", Endpoint: "grpcs://gateway-a:50051"})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	requestDone := make(chan struct{})
	go func() {
		defer close(requestDone)
		NewRouter(registry).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/openshell/gateway-a/api/v1/gateway", nil))
	}()
	<-started

	removeDone := make(chan error, 1)
	go func() {
		_, removeErr := registry.Remove(context.Background(), "gateway-a")
		removeDone <- removeErr
	}()

	select {
	case <-closed:
		t.Fatal("gateway closed before active request completed")
	case <-time.After(25 * time.Millisecond):
	}

	close(allowResponse)
	<-requestDone
	if err := <-removeDone; err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("gateway was not closed after request drained")
	}

	recorder := httptest.NewRecorder()
	NewRouter(registry).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/openshell/gateway-a/api/v1/gateway", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("removed gateway status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}
