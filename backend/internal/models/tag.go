package models

import "github.com/google/uuid"

// Tag maps core.tag — etichete disponibile pentru înregistrări domeniu.
type Tag struct {
	ID        uuid.UUID `json:"id"`
	Value     string    `json:"value"`
	SortOrder int       `json:"sort_order"`
}
