package services

import (
	"context"
	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
	"net"

	"github.com/google/uuid"
)

// DomainService defines the interface for domain business logic
type DomainService interface {
	SaveDomain(ctx context.Context, input models.SaveDomainInput) (*models.Domain, error)
	GetDomainByID(ctx context.Context, id uuid.UUID) (*models.Domain, error)
	GetDomains(ctx context.Context) ([]*models.Domain, error)
	WhitelistDomain(ctx context.Context, id uuid.UUID) error
	UpdateDomain(ctx context.Context, id uuid.UUID, input models.UpdateDomainInput) (*models.Domain, error)
}

type domainService struct {
	repo repository.DomainRepository
}

// NewDomainService creates a new domain service
func NewDomainService(repo repository.DomainRepository) DomainService {
	return &domainService{repo: repo}
}

func domainTypeFromValue(value string) string {
	if value == "" {
		return models.DomainTypeDomain
	}
	if net.ParseIP(value) != nil {
		return models.DomainTypeIP
	}
	return models.DomainTypeDomain
}

// SaveDomain: if no domain with same value+type exists, insert new domain and records; otherwise append input records to existing domain
func (s *domainService) SaveDomain(ctx context.Context, input models.SaveDomainInput) (*models.Domain, error) {
	typ := domainTypeFromValue(input.Value)
	existing, err := s.repo.GetByValueAndType(ctx, input.Value, typ)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		domain := &models.Domain{
			ID:        uuid.New(),
			Value:     input.Value,
			Type:      typ,
			Whitelist: input.Whitelist,
			Records:   nil,
		}
		for _, r := range input.Records {
			domain.Records = append(domain.Records, models.DomainRecord{
				ID:          uuid.New(),
				DomainID:    domain.ID,
				TicketID:    r.TicketID,
				Description: r.Description,
				Tags:        r.Tags,
				Date:        r.Date,
				Source:      r.Source,
			})
		}
		if err := s.repo.Insert(ctx, domain); err != nil {
			return nil, err
		}
		return domain, nil
	}

	// append new records to existing domain
	if len(input.Records) > 0 {
		newRecs := make([]models.DomainRecord, 0, len(input.Records))
		for _, r := range input.Records {
			newRecs = append(newRecs, models.DomainRecord{
				ID:          uuid.New(),
				DomainID:    existing.ID,
				TicketID:    r.TicketID,
				Description: r.Description,
				Tags:        r.Tags,
				Date:        r.Date,
				Source:      r.Source,
			})
		}
		if err := s.repo.InsertRecords(ctx, existing.ID, newRecs); err != nil {
			return nil, err
		}
		existing.Records = append(existing.Records, newRecs...)
	}
	return existing, nil
}

// GetDomainByID retrieves a domain by ID (with records)
func (s *domainService) GetDomainByID(ctx context.Context, id uuid.UUID) (*models.Domain, error) {
	return s.repo.GetByID(ctx, id)
}

// GetDomains retrieves all domains with their records
func (s *domainService) GetDomains(ctx context.Context) ([]*models.Domain, error) {
	return s.repo.List(ctx)
}

// WhitelistDomain sets whitelist = true for the domain
func (s *domainService) WhitelistDomain(ctx context.Context, id uuid.UUID) error {
	return s.repo.SetWhitelist(ctx, id, true)
}

// UpdateDomain updates only Value and/or Whitelist
func (s *domainService) UpdateDomain(ctx context.Context, id uuid.UUID, input models.UpdateDomainInput) (*models.Domain, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Value != nil {
		current.Value = *input.Value
		current.Type = domainTypeFromValue(current.Value)
	}
	if input.Whitelist != nil {
		current.Whitelist = *input.Whitelist
	}
	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}
