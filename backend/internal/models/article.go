package models

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Article struct {
	ID         string `json:"id" db:"id"`
	Provider   string `json:"provider" db:"provider"`
	ExternalID int64  `json:"external_id" db:"external_id"`
	Type       string `json:"type" db:"type"`

	Title       string  `json:"title" db:"title"`
	Description string  `json:"description" db:"description"`
	Summary     string  `json:"summary" db:"summary"`
	Body        *string `json:"body,omitempty" db:"body"`

	Language     string `json:"language" db:"language"`
	CanonicalURL string `json:"canonical_url" db:"canonical_url"`
	HotlinkURL   string `json:"hotlink_url" db:"hotlink_url"`
	ImageURL     string `json:"image_url" db:"image_url"`

	PublishedAt       time.Time `json:"published_at" db:"published_at"`
	ExternalUpdatedAt time.Time `json:"external_updated_at" db:"external_updated_at"`

	ContentHash string `json:"content_hash" db:"content_hash"`

	SyncStatus   string     `json:"sync_status" db:"sync_status"`
	SyncAttempts int        `json:"sync_attempts" db:"sync_attempts"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty" db:"last_synced_at"`
	SyncError    *string    `json:"sync_error,omitempty" db:"sync_error"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CentralArticleDTO is the payload sent to the central management system.
// It exposes only the fields relevant for the CMS, not internal sync/DB details.
type CentralArticleDTO struct {
	ID           string  `json:"id"`
	Provider     string  `json:"provider"`
	ExternalID   int64   `json:"external_id"`
	Type         string  `json:"type"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	Summary      string  `json:"summary"`
	Body         *string `json:"body,omitempty"`
	Language     string  `json:"language"`
	CanonicalURL string  `json:"canonical_url"`
	HotlinkURL   string  `json:"hotlink_url"`
	ImageURL     string  `json:"image_url"`
	PublishedAt  string  `json:"published_at"` // RFC3339
}

// ArticleToCentralDTO maps an internal Article to the DTO sent to the central system.
func ArticleToCentralDTO(a Article) CentralArticleDTO {
	return CentralArticleDTO{
		ID:           a.ID,
		Provider:     a.Provider,
		ExternalID:   a.ExternalID,
		Type:         a.Type,
		Title:        a.Title,
		Description:  a.Description,
		Summary:      a.Summary,
		Body:         a.Body,
		Language:     a.Language,
		CanonicalURL: a.CanonicalURL,
		HotlinkURL:   a.HotlinkURL,
		ImageURL:     a.ImageURL,
		PublishedAt:  a.PublishedAt.Format(time.RFC3339),
	}
}

type ExternalArticleDTO struct {
	ID           int64   `json:"id"`
	Type         string  `json:"type"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	Summary      string  `json:"summary"`
	Body         *string `json:"body"`
	Language     string  `json:"language"`
	CanonicalURL string  `json:"canonicalUrl"`
	HotlinkURL   string  `json:"hotlinkUrl"`
	ImageURL     string  `json:"imageUrl"`
	Date         string  `json:"date"`
	LastModified int64   `json:"lastModified"`

	LeadMedia *struct {
		ImageURL string `json:"imageUrl"`
	} `json:"leadMedia"`
}

func MapToArticle(dto ExternalArticleDTO) (Article, error) {
	publishedAt, err := time.Parse(time.RFC3339, dto.Date)
	if err != nil {
		return Article{}, err
	}

	externalUpdatedAt := time.UnixMilli(dto.LastModified)

	imageURL := dto.ImageURL
	if imageURL == "" && dto.LeadMedia != nil {
		imageURL = dto.LeadMedia.ImageURL
	}

	now := time.Now()

	article := Article{
		ID:                uuid.NewString(),
		Provider:          "pulselive",
		ExternalID:        dto.ID,
		Type:              dto.Type,
		Title:             dto.Title,
		Description:       dto.Description,
		Summary:           dto.Summary,
		Body:              dto.Body,
		Language:          dto.Language,
		CanonicalURL:      dto.CanonicalURL,
		HotlinkURL:        dto.HotlinkURL,
		ImageURL:          imageURL,
		PublishedAt:       publishedAt,
		ExternalUpdatedAt: externalUpdatedAt,
		ContentHash:       ComputeHash(dto),
		SyncStatus:        "pending",
		SyncAttempts:      0,
		LastSyncedAt:      nil,
		SyncError:         nil,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	return article, nil
}

func ComputeHash(dto ExternalArticleDTO) string {
	body := ""
	if dto.Body != nil {
		body = *dto.Body
	}

	imageURL := dto.ImageURL
	if imageURL == "" && dto.LeadMedia != nil {
		imageURL = dto.LeadMedia.ImageURL
	}

	data := strings.Join([]string{
		dto.Title,
		dto.Description,
		dto.Summary,
		body,
		dto.CanonicalURL,
		dto.HotlinkURL,
		imageURL,
		dto.Date,
	}, "|")

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
