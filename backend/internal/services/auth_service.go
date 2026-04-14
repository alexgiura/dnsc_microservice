package services

import (
	"context"
	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AuthService handles login, session validation, logout.
type AuthService interface {
	Login(ctx context.Context, username, password string) (*models.UserPublic, uuid.UUID, time.Time, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	ValidateSession(ctx context.Context, sessionID uuid.UUID) (*models.AppUser, error)
	ToPublic(u *models.AppUser) models.UserPublic
}

type authService struct {
	repo   repository.AuthRepository
	ttl    time.Duration
}

// NewAuthService creates auth service. ttl is session lifetime (e.g. 7 days).
func NewAuthService(repo repository.AuthRepository, sessionTTL time.Duration) AuthService {
	if sessionTTL <= 0 {
		sessionTTL = 7 * 24 * time.Hour
	}
	return &authService{repo: repo, ttl: sessionTTL}
}

func (s *authService) ToPublic(u *models.AppUser) models.UserPublic {
	if u == nil {
		return models.UserPublic{}
	}
	return models.UserPublic{ID: u.ID, Username: u.Username}
}

func (s *authService) Login(ctx context.Context, username, password string) (*models.UserPublic, uuid.UUID, time.Time, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return nil, uuid.Nil, time.Time{}, fmt.Errorf("username and password are required")
	}
	u, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, uuid.Nil, time.Time{}, err
	}
	if u == nil || u.Password != password {
		return nil, uuid.Nil, time.Time{}, errors.New("invalid credentials")
	}
	if !u.IsActive {
		return nil, uuid.Nil, time.Time{}, errors.New("user inactive")
	}
	sessionID := uuid.New()
	expiresAt := time.Now().UTC().Add(s.ttl)
	if err := s.repo.InsertSession(ctx, sessionID, u.ID, expiresAt); err != nil {
		return nil, uuid.Nil, time.Time{}, err
	}
	pub := s.ToPublic(u)
	return &pub, sessionID, expiresAt, nil
}

func (s *authService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.repo.DeleteSession(ctx, sessionID)
}

func (s *authService) ValidateSession(ctx context.Context, sessionID uuid.UUID) (*models.AppUser, error) {
	sess, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, errors.New("session not found")
	}
	if sess.ExpiresAt.Before(time.Now()) {
		_ = s.repo.DeleteSession(ctx, sessionID)
		return nil, errors.New("session expired")
	}
	u, err := s.repo.GetUserByID(ctx, sess.UserID)
	if err != nil {
		return nil, err
	}
	if u == nil || !u.IsActive {
		return nil, errors.New("user invalid")
	}
	return u, nil
}
