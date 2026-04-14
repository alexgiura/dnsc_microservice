package services

import (
	"context"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// TagService exposes tag catalog for API.
type TagService interface {
	ListTags(ctx context.Context) ([]models.Tag, error)
}

type tagService struct {
	repo repository.TagRepository
}

func NewTagService(repo repository.TagRepository) TagService {
	return &tagService{repo: repo}
}

func (s *tagService) ListTags(ctx context.Context) ([]models.Tag, error) {
	return s.repo.ListTags(ctx)
}
