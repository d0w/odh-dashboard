package fleet

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testConfig struct{ name string }

type testInstance struct {
	handler http.Handler
	closed  *int
}

func (i testInstance) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	i.handler.ServeHTTP(w, r)
}

func (i testInstance) Close() error {
	(*i.closed)++
	return nil
}

func TestRegistryForwardsPathAfterInstanceID(t *testing.T) {
	var gotPath, gotRawPath, gotQuery, gotRequestURI string
	registry := NewRegistry(func(_ context.Context, _ testConfig) (testInstance, error) {
		return testInstance{
			handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				gotPath = req.URL.Path
				gotRawPath = req.URL.RawPath
				gotQuery = req.URL.RawQuery
				gotRequestURI = req.RequestURI
				w.WriteHeader(http.StatusNoContent)
			}),
		}, nil
	}, func(config testConfig) string { return config.name })

	created, err := registry.Register(context.Background(), testConfig{name: "openshell"})
	if err != nil {
		t.Fatalf("register instance: %v", err)
	}
	if !created {
		t.Fatal("instance was not registered")
	}

	router, err := NewRouter("/agent-ops/api/openshell", registry)
	if err != nil {
		t.Fatalf("create router: %v", err)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(
		http.MethodGet,
		"/agent-ops/api/openshell/openshell/v1/models%2Fmy-model?include=details",
		nil,
	))

	if recorder.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if gotPath != "/v1/models/my-model" {
		t.Errorf("path = %q, want %q", gotPath, "/v1/models/my-model")
	}
	if gotRawPath != "/v1/models%2Fmy-model" {
		t.Errorf("raw path = %q, want %q", gotRawPath, "/v1/models%2Fmy-model")
	}
	if gotQuery != "include=details" {
		t.Errorf("query = %q, want %q", gotQuery, "include=details")
	}
	if gotRequestURI != "/v1/models%2Fmy-model?include=details" {
		t.Errorf("request URI = %q, want %q", gotRequestURI, "/v1/models%2Fmy-model?include=details")
	}
}

func TestRegistrySupportsArbitraryTypes(t *testing.T) {
	closed := 0
	registry := NewRegistry(func(_ context.Context, _ testConfig) (testInstance, error) {
		return testInstance{
			handler: http.NotFoundHandler(),
			closed:  &closed,
		}, nil
	}, func(config testConfig) string { return config.name })

	created, err := registry.Register(context.Background(), testConfig{name: "example"})
	if err != nil {
		t.Fatalf("register entry: %v", err)
	}
	if !created {
		t.Fatal("entry was not registered")
	}

	entry, found := registry.Get("example")
	if !found {
		t.Fatal("entry was not found")
	}
	if entry.closed != &closed {
		t.Error("registry returned unexpected instance")
	}

	if err := registry.Close(context.Background()); err != nil {
		t.Fatalf("close registry: %v", err)
	}
	if closed != 1 {
		t.Errorf("closed = %d, want 1", closed)
	}
}
