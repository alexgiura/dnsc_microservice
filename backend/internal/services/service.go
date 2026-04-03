package services

import (
	"dnsc_microservice/internal/clients/rtir"
	"dnsc_microservice/internal/repository"
)

// AppServices holds all service interfaces
type AppServices struct {
	Domain DomainService
}

// NewAppServices initializes all services
func NewAppServices(repos *repository.Repository, rtirClient *rtir.Client, rtirSyncTimezone string, rtirOverlapMinutes int) *AppServices {
	return &AppServices{
		Domain: NewDomainService(repos.Domain, rtirClient, rtirSyncTimezone, rtirOverlapMinutes),
	}
}
