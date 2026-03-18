package external

import (
	"context"

	"cortex/internal/models"
)

// ArticlesClient is a generic abstraction over any external
type ArticlesClient interface {
	FetchLatest(ctx context.Context, page int, pageSize int) ([]models.ExternalArticleDTO, error)
}
