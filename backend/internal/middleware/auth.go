package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"

	"github.com/google/uuid"
)

type ctxKey int

const userCtxKey ctxKey = 1

// UserFromContext returns the authenticated app user set by AuthMiddleware.
func UserFromContext(ctx context.Context) (*models.AppUser, bool) {
	u, ok := ctx.Value(userCtxKey).(*models.AppUser)
	return u, ok
}

// SessionValidator is implemented by services.AuthService for session checks.
type SessionValidator interface {
	ValidateSession(ctx context.Context, sessionID uuid.UUID) (*models.AppUser, error)
}

func isPublicPath(path, method string) bool {
	if method == http.MethodOptions {
		return true
	}
	if path == "/healthz" || path == "/health" {
		return method == http.MethodGet
	}
	if path == "/auth/login" && method == http.MethodPost {
		return true
	}
	if path == "/api/public/domains" && method == http.MethodGet {
		return true
	}
	return false
}

// AuthMiddleware enforces session cookie for protected routes.
func AuthMiddleware(validator SessionValidator, cookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublicPath(r.URL.Path, r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie(cookieName)
			if err != nil || strings.TrimSpace(cookie.Value) == "" {
				httpUnauthorized(w)
				return
			}
			sid, err := uuid.Parse(cookie.Value)
			if err != nil {
				httpUnauthorized(w)
				return
			}
			user, err := validator.ValidateSession(r.Context(), sid)
			if err != nil || user == nil {
				httpUnauthorized(w)
				return
			}
			ctx := context.WithValue(r.Context(), userCtxKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func httpUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": "UNAUTHORIZED", "message": "Unauthorized"})
}
