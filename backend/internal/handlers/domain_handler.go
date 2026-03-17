package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type DomainHandler struct {
	domain services.DomainService
}

func NewDomainHandler(domain services.DomainService) *DomainHandler {
	return &DomainHandler{domain: domain}
}

func (h *DomainHandler) SaveDomain(w http.ResponseWriter, r *http.Request) {
	var input models.SaveDomainInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}

	if input.Value == "" {
		respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, "value is required", "")
		return
	}

	domain, err := h.domain.SaveDomain(r.Context(), input)
	if err != nil {
		statusCode, code, message := parseDatabaseError(err)
		respondWithError(w, statusCode, code, message, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(domain); err != nil {
		log.Printf("Error encoding domain response: %v", err)
	}
}

// GetDomainByID handles GET /api/domains/{id}
func (h *DomainHandler) GetDomainByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "id is required", "")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid id format", err.Error())
		return
	}

	domain, err := h.domain.GetDomainByID(r.Context(), id)
	if err != nil {
		statusCode, code, message := parseDatabaseError(err)
		respondWithError(w, statusCode, code, message, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(domain); err != nil {
		log.Printf("Error encoding domain response: %v", err)
	}
}

// GetDomains handles GET /api/domains
func (h *DomainHandler) GetDomains(w http.ResponseWriter, r *http.Request) {
	domains, err := h.domain.GetDomains(r.Context())
	if err != nil {
		statusCode, code, message := parseDatabaseError(err)
		respondWithError(w, statusCode, code, message, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(domains); err != nil {
		log.Printf("Error encoding domains response: %v", err)
	}
}

// WhitelistDomain handles POST /api/domains/{id}/whitelist
// It marks a domain as no longer a threat (whitelist = true)
func (h *DomainHandler) WhitelistDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "id is required", "")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid id format", err.Error())
		return
	}

	if err := h.domain.WhitelistDomain(r.Context(), id); err != nil {
		statusCode, code, message := parseDatabaseError(err)
		respondWithError(w, statusCode, code, message, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateDomain handles PATCH /api/domains/{id}
// Only provided fields are updated; id, count, first_seen, last_seen are never modified from the request.
func (h *DomainHandler) UpdateDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "id is required", "")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid id format", err.Error())
		return
	}

	var input models.UpdateDomainInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}

	updated, err := h.domain.UpdateDomain(r.Context(), id, input)
	if err != nil {
		statusCode, code, message := parseDatabaseError(err)
		respondWithError(w, statusCode, code, message, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updated); err != nil {
		log.Printf("Error encoding updated domain response: %v", err)
	}
}
