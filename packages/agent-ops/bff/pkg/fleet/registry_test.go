package fleet

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegistryForwardsPathAfterGatewayID(t *testing.T) {
	var gotPath, gotRawPath, gotQuery, gotRequestURI string
	registry := NewRegistry(func(_ context.Context, _ GatewayConfig) (GatewayInstance, error) {
		return GatewayInstance{
			handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				gotPath = req.URL.Path
				gotRawPath = req.URL.RawPath
				gotQuery = req.URL.RawQuery
				gotRequestURI = req.RequestURI
				w.WriteHeader(http.StatusNoContent)
			}),
		}, nil
	})

	created, err := registry.Register(context.Background(), GatewayConfig{ID: "openshell"})
	if err != nil {
		t.Fatalf("register gateway: %v", err)
	}
	if !created {
		t.Fatal("gateway was not registered")
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
