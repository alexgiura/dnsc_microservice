package routes

import (
	"cortex/internal/handlers"

	"github.com/gorilla/mux"
)

func RegisterRoutes() *mux.Router {
	router := mux.NewRouter()

	handlers.RegisterSystemRoutes(router)

	return router
}
