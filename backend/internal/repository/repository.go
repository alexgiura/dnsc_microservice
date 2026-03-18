package repository

import "github.com/jackc/pgx/v4/pgxpool"

// Repository holds all repository interfaces
type Repository struct {
	Article ArticleRepository
}

// NewRepository initializes all repositories
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		Article: NewArticleRepository(db),
	}
}
