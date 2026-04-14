package models

import (
	"time"

	"github.com/google/uuid"
)

// AppUser maps core.app_user (password stored as plain text per project policy).
type AppUser struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserPublic is returned from /auth/login and /auth/me (no password).
type UserPublic struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

// UserSession maps core.user_session.
type UserSession struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
	CreatedAt time.Time
}

// LoginRequest body for POST /auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse body after successful login.
type LoginResponse struct {
	User UserPublic `json:"user"`
}

// MeResponse body for GET /auth/me.
type MeResponse struct {
	User UserPublic `json:"user"`
}
