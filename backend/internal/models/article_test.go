package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMapToArticle_Success(t *testing.T) {
	body := "full body"
	dto := ExternalArticleDTO{
		ID:           123,
		Type:         "news",
		Title:        "Some title",
		Description:  "desc",
		Summary:      "summary",
		Body:         &body,
		Language:     "en",
		CanonicalURL: "https://example.com/article",
		HotlinkURL:   "https://example.com/hot",
		ImageURL:     "",
		Date:         "2024-01-02T15:04:05Z",
		LastModified: time.Date(2024, 1, 2, 16, 0, 0, 0, time.UTC).UnixMilli(),
		LeadMedia: &struct {
			ImageURL string `json:"imageUrl"`
		}{
			ImageURL: "https://example.com/lead.jpg",
		},
	}

	a, err := MapToArticle(dto)
	require.NoError(t, err)

	require.Equal(t, "pulselive", a.Provider)
	require.Equal(t, dto.ID, a.ExternalID)
	require.Equal(t, dto.Type, a.Type)
	require.Equal(t, dto.Title, a.Title)
	require.Equal(t, dto.Description, a.Description)
	require.Equal(t, dto.Summary, a.Summary)
	require.Equal(t, dto.Language, a.Language)
	require.Equal(t, dto.CanonicalURL, a.CanonicalURL)
	require.Equal(t, dto.HotlinkURL, a.HotlinkURL)

	// ImageURL should fall back to leadMedia.imageUrl when ImageURL is empty.
	require.Equal(t, "https://example.com/lead.jpg", a.ImageURL)

	require.NotEmpty(t, a.ContentHash)
	require.Equal(t, "pending", a.SyncStatus)
	require.Equal(t, 0, a.SyncAttempts)

	// Timestamps: exact mapping for ExternalUpdatedAt from LastModified.
	require.False(t, a.PublishedAt.IsZero())
	require.Equal(t, time.UnixMilli(dto.LastModified), a.ExternalUpdatedAt)
	require.False(t, a.CreatedAt.IsZero())
	require.False(t, a.UpdatedAt.IsZero())
}

func TestMapToArticle_ImageURL_Priority(t *testing.T) {
	// When ImageURL is set, it is used; LeadMedia is not used.
	dto := ExternalArticleDTO{
		ImageURL: "https://main.jpg",
		LeadMedia: &struct {
			ImageURL string `json:"imageUrl"`
		}{
			ImageURL: "https://lead.jpg",
		},
		Date: "2024-01-02T15:04:05Z",
	}

	a, err := MapToArticle(dto)
	require.NoError(t, err)
	require.Equal(t, "https://main.jpg", a.ImageURL)
}

func TestMapToArticle_InvalidDate(t *testing.T) {
	dto := ExternalArticleDTO{
		ID:   1,
		Date: "invalid-date",
	}

	_, err := MapToArticle(dto)
	require.Error(t, err)
}

func TestComputeHash_StableForSameInput(t *testing.T) {
	dto := ExternalArticleDTO{
		Title:       "t",
		Description: "d",
		Summary:     "s",
		Body:        strPtr("body"),
		Language:    "en",
		CanonicalURL:"https://c",
		HotlinkURL:  "https://h",
		ImageURL:    "https://i",
		Date:        "2024-01-02T15:04:05Z",
	}

	h1 := ComputeHash(dto)
	h2 := ComputeHash(dto)

	require.Equal(t, h1, h2)
}

func TestComputeHash_ChangesWhenContentChanges(t *testing.T) {
	dto1 := ExternalArticleDTO{
		Title:       "title1",
		Description: "d",
		Summary:     "s",
		Body:        strPtr("body"),
		CanonicalURL:"https://c",
		HotlinkURL:  "https://h",
		ImageURL:    "https://i",
		Date:        "2024-01-02T15:04:05Z",
	}
	dto2 := dto1
	dto2.Title = "title2"

	h1 := ComputeHash(dto1)
	h2 := ComputeHash(dto2)

	require.NotEqual(t, h1, h2)
}

func TestComputeHash_BodyNilDoesNotPanic(t *testing.T) {
	dto := ExternalArticleDTO{
		Title:       "t",
		Description: "d",
		Summary:     "s",
		Body:        nil,
		CanonicalURL:"https://c",
		HotlinkURL:  "https://h",
		ImageURL:    "https://i",
		Date:        "2024-01-02T15:04:05Z",
	}

	_ = ComputeHash(dto) // should not panic
}

func strPtr(s string) *string { return &s }

