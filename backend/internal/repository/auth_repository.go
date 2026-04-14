package repository

import (
	"context"
	"dnsc_microservice/internal/models"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// AuthRepository persists app users and sessions.
type AuthRepository interface {
	GetUserByUsername(ctx context.Context, username string) (*models.AppUser, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.AppUser, error)
	InsertSession(ctx context.Context, sessionID, userID uuid.UUID, expiresAt time.Time) error
	GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*models.UserSession, error)
	DeleteSession(ctx context.Context, sessionID uuid.UUID) error
}

type authRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) GetUserByUsername(ctx context.Context, username string) (*models.AppUser, error) {
	username = strings.TrimSpace(username)
	row := r.db.QueryRow(ctx, `
		SELECT id, username, password, is_active, created_at, updated_at
		FROM core.app_user WHERE username = $1
	`, username)
	var u models.AppUser
	if err := row.Scan(&u.ID, &u.Username, &u.Password, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return &u, nil
}

func (r *authRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.AppUser, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, username, password, is_active, created_at, updated_at
		FROM core.app_user WHERE id = $1
	`, id)
	var u models.AppUser
	if err := row.Scan(&u.ID, &u.Username, &u.Password, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

func (r *authRepository) InsertSession(ctx context.Context, sessionID, userID uuid.UUID, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO core.user_session (id, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, sessionID, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (r *authRepository) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*models.UserSession, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, expires_at, created_at FROM core.user_session WHERE id = $1
	`, sessionID)
	var s models.UserSession
	if err := row.Scan(&s.ID, &s.UserID, &s.ExpiresAt, &s.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &s, nil
}

func (r *authRepository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM core.user_session WHERE id = $1`, sessionID)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
