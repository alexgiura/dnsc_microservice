package repository

import (
	"github.com/jackc/pgx/v4/pgxpool"
)

// Repository holds all repository interfaces
type Repository struct {
	Domain DomainRepository
	Auth   AuthRepository
	Tag    TagRepository
}

// NewRepository initializes all repositories
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		Domain: NewDomainRepository(db),
		Auth:   NewAuthRepository(db),
		Tag:    NewTagRepository(db),
	}
}
