package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

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

// GetPublicBlacklistedDomains handles GET /api/public/domains
// and returns only public-safe blacklisted domains (value/type/date).
func (h *DomainHandler) GetPublicBlacklistedDomains(w http.ResponseWriter, r *http.Request) {
	items, err := h.domain.GetPublicBlacklistedDomains(r.Context())
	if err != nil {
		statusCode, code, message := parseDatabaseError(err)
		respondWithError(w, statusCode, code, message, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Printf("Error encoding public domains response: %v", err)
	}
}

// WhitelistDomain handles POST /api/domains/{id}/whitelist
// It updates core.domains.whitelist and inserts a row into core.domain_status.
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

	var input models.WhitelistDomainInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}

	if input.Whitelist == nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, "whitelist is required", "")
		return
	}
	if input.ChangeBy == "" {
		respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, "changeBy is required", "")
		return
	}
	if input.DomainID != nil && *input.DomainID != id {
		respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, "domainId does not match path id", "")
		return
	}

	notes := ""
	if input.Notes != nil {
		notes = *input.Notes
	}

	if err := h.domain.ChangeDomainStatus(r.Context(), id, *input.Whitelist, input.ChangeBy, notes); err != nil {
		statusCode, code, message := parseDatabaseError(err)
		respondWithError(w, statusCode, code, message, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RequestWhitelist handles POST /api/domains/{id}/whitelist-requests
// and creates one request in core.whitelist_requests.
func (h *DomainHandler) RequestWhitelist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "id is required", "")
		return
	}

	domainID, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid id format", err.Error())
		return
	}

	var input models.CreateWhitelistRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}

	if strings.TrimSpace(input.FirstName) == "" {
		respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, "first_name is required", "")
		return
	}
	if strings.TrimSpace(input.LastName) == "" {
		respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, "last_name is required", "")
		return
	}
	if strings.TrimSpace(input.Email) == "" {
		respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, "email is required", "")
		return
	}
	if strings.TrimSpace(input.Address) == "" {
		respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, "address is required", "")
		return
	}
	if strings.TrimSpace(input.Phone) == "" {
		respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, "phone is required", "")
		return
	}
	if strings.TrimSpace(input.Reason) == "" {
		respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, "reason is required", "")
		return
	}

	request, err := h.domain.RequestWhitelist(r.Context(), domainID, input)
	if err != nil {
		statusCode, code, message := parseDatabaseError(err)
		respondWithError(w, statusCode, code, message, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(request); err != nil {
		log.Printf("Error encoding whitelist request response: %v", err)
	}
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
