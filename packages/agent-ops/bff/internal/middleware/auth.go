package middleware

// TODO: Eventually this will need to be changed to implement the upstream auth middleware's interface

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Gkrumbach07/openshell-dashboard/backend/pkg/auth"
)

type contextKey int

const (
	tokenContextKey contextKey = iota
	userContextKey
)

type AuthReplacer struct {
	*auth.Middleware
	cfg auth.Config
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, `{"code":"unauthorized","message":%q}`, message)
}

func NewAuthReplacer(cfg auth.Config) *AuthReplacer {
	return &AuthReplacer{
		Middleware: &auth.Middleware{},
		cfg:        cfg,
	}
}

func (m *AuthReplacer) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.cfg.Disabled {
			ctx := context.WithValue(r.Context(), userContextKey, "dev-user")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		token := r.Header.Get(m.cfg.TokenHeader)
		if token == "" {
			if bearer := r.Header.Get("Authorization"); strings.HasPrefix(bearer, "Bearer ") {
				token = strings.TrimPrefix(bearer, "Bearer ")
			}
		}
		if token == "" {
			writeUnauthorized(w, "not authenticated")
			return
		}

		ctx := context.WithValue(r.Context(), tokenContextKey, token)
		if user := r.Header.Get(m.cfg.UserHeader); user != "" {
			ctx = context.WithValue(ctx, userContextKey, user)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
