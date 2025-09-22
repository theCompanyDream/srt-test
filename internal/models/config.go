package models

import "time"

type ValidationError struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// LangResponse represents the response from the language detection endpoint
type LangResponse struct {
	Lang string `json:"lang"`
}

// CaptionEntry represents a single caption with timing
type CaptionEntry struct {
	StartTime time.Duration
	EndTime   time.Duration
	Text      string
}

// Config holds the program configuration
type Config struct {
	FilePath string
	TStart   time.Duration
	TEnd     time.Duration
	Coverage float64
	Endpoint string
}
