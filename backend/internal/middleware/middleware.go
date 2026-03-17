package middleware

import (
	"net/http"
	"strings"

	"github.com/rs/cors"
)

// CorsMiddleware handles CORS for all requests
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{http.MethodPost, http.MethodGet, http.MethodPatch, http.MethodDelete, http.MethodPut},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: false,
		})

		handler := c.Handler(next)
		handler.ServeHTTP(w, r)
	})
}

// APIKeyMiddleware validates API key from X-API-Key header
// Add this middleware to routes that need API key authentication
func APIKeyMiddleware(validAPIKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				apiKey = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			}

			if apiKey == "" || apiKey != validAPIKey {
				http.Error(w, "Invalid or missing API key", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
