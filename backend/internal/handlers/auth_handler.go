package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"dnsc_microservice/internal/config"
	"dnsc_microservice/internal/middleware"
	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/google/uuid"
)

type AuthHandler struct {
	auth services.AuthService
	cfg  *config.Config
	ttl  time.Duration
}

func NewAuthHandler(auth services.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		auth: auth,
		cfg:  cfg,
		ttl:  cfg.SessionTTL(),
	}
}

// Login POST /auth/login — sets HttpOnly session cookie.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var in models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	user, sessionID, expiresAt, err := h.auth.Login(r.Context(), in.Username, in.Password)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "invalid") || strings.Contains(msg, "inactive") || strings.Contains(msg, "required") {
			respondWithError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Login failed", "")
			return
		}
		respondWithError(w, http.StatusInternalServerError, ErrCodeInternalError, "Login failed", err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.SessionCookieName,
		Value:    sessionID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.SessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.ttl.Seconds()),
		Expires:  expiresAt,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(models.LoginResponse{User: *user})
}

// Logout POST /auth/logout — clears session row and cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.cfg.SessionCookieName)
	if err != nil || cookie.Value == "" {
		clearSessionCookie(w, h.cfg)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	sid, err := uuid.Parse(cookie.Value)
	if err != nil {
		clearSessionCookie(w, h.cfg)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	_ = h.auth.Logout(r.Context(), sid)
	clearSessionCookie(w, h.cfg)
	w.WriteHeader(http.StatusNoContent)
}

func clearSessionCookie(w http.ResponseWriter, cfg *config.Config) {
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.SessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

// Me GET /auth/me — user must be authenticated (middleware).
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok || u == nil {
		respondWithError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Unauthorized", "")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(models.MeResponse{User: h.auth.ToPublic(u)})
}
