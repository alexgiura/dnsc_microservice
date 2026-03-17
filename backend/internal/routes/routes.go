package routes

import (
	"dnsc_microservice/internal/handlers"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

func RegisterRoutes(appServices *services.AppServices) *mux.Router {
	router := mux.NewRouter()

	handlers.RegisterSystemRoutes(router)

	domainHandler := handlers.NewDomainHandler(appServices.Domain)
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/domains", domainHandler.SaveDomain).Methods("POST")
	apiRouter.HandleFunc("/domains", domainHandler.GetDomains).Methods("GET")
	apiRouter.HandleFunc("/domains/{id}", domainHandler.GetDomainByID).Methods("GET")
	apiRouter.HandleFunc("/domains/{id}", domainHandler.UpdateDomain).Methods("PATCH")
	apiRouter.HandleFunc("/domains/{id}/whitelist", domainHandler.WhitelistDomain).Methods("POST")

	return router
}
