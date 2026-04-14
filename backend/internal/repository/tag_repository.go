package repository

import (
	"context"
	"fmt"

	"dnsc_microservice/internal/models"

	"github.com/jackc/pgx/v4/pgxpool"
)

// TagRepository reads tag definitions from core.tag.
type TagRepository interface {
	ListTags(ctx context.Context) ([]models.Tag, error)
}

type tagRepository struct {
	db *pgxpool.Pool
}

func NewTagRepository(db *pgxpool.Pool) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) ListTags(ctx context.Context) ([]models.Tag, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, value, sort_order
		FROM core.tag
		ORDER BY sort_order ASC, value ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()

	var out []models.Tag
	for rows.Next() {
		var t models.Tag
		if err := rows.Scan(&t.ID, &t.Value, &t.SortOrder); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows tags: %w", err)
	}
	return out, nil
}
