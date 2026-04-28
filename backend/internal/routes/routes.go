package routes

import (
	"net/http"

	"dnsc_microservice/internal/config"
	"dnsc_microservice/internal/handlers"
	"dnsc_microservice/internal/middleware"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

// RegisterRoutes builds the HTTP handler with auth + CORS.
func RegisterRoutes(appServices *services.AppServices, cfg *config.Config) http.Handler {
	router := mux.NewRouter()

	handlers.RegisterSystemRoutes(router)

	authHandler := handlers.NewAuthHandler(appServices.Auth, cfg)
	router.HandleFunc("/auth/login", authHandler.Login).Methods("POST")

	domainHandler := handlers.NewDomainHandler(appServices.Domain)
	tagHandler := handlers.NewTagHandler(appServices.Tag)
	router.HandleFunc("/api/public/domains", domainHandler.GetPublicBlacklistedDomains).Methods("GET")

	router.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST")
	router.HandleFunc("/auth/me", authHandler.Me).Methods("GET")

	router.HandleFunc("/api/dashboard", domainHandler.GetDashboard).Methods("GET")
	router.HandleFunc("/api/domains", domainHandler.SaveDomain).Methods("POST")
	router.HandleFunc("/api/domains", domainHandler.GetDomains).Methods("GET")
	router.HandleFunc("/api/domains/{id}", domainHandler.GetDomainByID).Methods("GET")
	router.HandleFunc("/api/domains/{id}", domainHandler.UpdateDomain).Methods("PATCH")
	router.HandleFunc("/api/domains/{id}/whitelist", domainHandler.WhitelistDomain).Methods("POST")
	router.HandleFunc("/api/domains/{id}/whitelist-requests", domainHandler.RequestWhitelist).Methods("POST")
	router.HandleFunc("/api/rtir/import-errors", domainHandler.GetRTIRImportErrors).Methods("GET")
	router.HandleFunc("/api/rtir/tickets/{ticketId}/reimport", domainHandler.ReimportRTIRTicket).Methods("POST")
	router.HandleFunc("/api/rtir/domain-records/sync-created-dates", domainHandler.SyncDomainRecordsDatesFromRTIRCreated).Methods("POST")
	router.HandleFunc("/api/tags", tagHandler.ListTags).Methods("GET")

	withAuth := middleware.AuthMiddleware(appServices.Auth, cfg.SessionCookieName)(router)
	return middleware.CorsMiddleware(cfg.CORSOriginsList())(withAuth)
}
