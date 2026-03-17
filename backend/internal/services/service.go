package services

import (
	"dnsc_microservice/internal/repository"
)

// AppServices holds all service interfaces
type AppServices struct {
	Domain DomainService
}

// NewAppServices initializes all services
func NewAppServices(repos *repository.Repository) *AppServices {
	return &AppServices{
		Domain: NewDomainService(repos.Domain),
	}
}
