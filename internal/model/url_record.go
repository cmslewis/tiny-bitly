package model

import "time"

// URLRecord is the structure for use in code.
type URLRecord struct {
	OriginalURL string    `json:"originalUrl"`
	ShortCode   string    `json:"shortCode"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

// URLRecordEntity will be stored as a row in the database.
type URLRecordEntity struct {
	Entity
	URLRecord
}
