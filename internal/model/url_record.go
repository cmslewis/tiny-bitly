package model

// URLRecord is the structure for use in code.
type URLRecord struct {
	OriginalURL string `json:"originalUrl"`
	ShortCode   string `json:"shortCode"`
}

// URLRecordEntity will be stored as a row in the database.
type URLRecordEntity struct {
	Entity
	URLRecord
}
