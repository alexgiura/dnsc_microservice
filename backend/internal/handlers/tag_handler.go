package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"
)

type TagHandler struct {
	tag services.TagService
}

func NewTagHandler(tag services.TagService) *TagHandler {
	return &TagHandler{tag: tag}
}

// ListTags GET /api/tags
func (h *TagHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.tag.ListTags(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, ErrCodeInternalError, "failed to list tags", err.Error())
		return
	}
	if tags == nil {
		tags = []models.Tag{}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string][]models.Tag{"tags": tags}); err != nil {
		log.Printf("encode tags: %v", err)
	}
}
