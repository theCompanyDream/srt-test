package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/theCompanyDream/srt-test/internal/cmd"
	"github.com/theCompanyDream/srt-test/internal/models"
	"github.com/theCompanyDream/srt-test/internal/parse"
)

func main() {
	config, err := cmd.ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Validate file type
	if !isValidFileType(config.FilePath) {
		os.Exit(1)
	}

	// Parse caption file
	captions, err := parse.ParseCaptionFile(config.FilePath)
	if err != nil {
		printValidationError("file_parse_error", fmt.Sprintf("Failed to parse caption file: %v", err))
		os.Exit(0)
	}

	var validationErrors []models.ValidationError

	// Validate coverage
	if !validateCoverage(captions, config.TStart, config.TEnd, config.Coverage) {
		validationErrors = append(validationErrors, models.ValidationError{
			Type:        "insufficient_coverage",
			Description: fmt.Sprintf("Captions do not cover required %.1f%% of time range %v to %v", config.Coverage*100, config.TStart, config.TEnd),
		})
	}

	// Extract and validate language
	allText := parse.ExtractAllText(captions)
	if !validateLanguage(allText, config.Endpoint) {
		validationErrors = append(validationErrors, models.ValidationError{
			Type:        "invalid_language",
			Description: "Caption language is not en-US or language detection failed",
		})
	}

	// Print validation errors
	for _, err := range validationErrors {
		printValidationError(err.Type, err.Description)
	}

	os.Exit(0)
}

func isValidFileType(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".vtt" || ext == ".srt"
}

func validateCoverage(captions []models.CaptionEntry, tStart, tEnd time.Duration, requiredCoverage float64) bool {
	totalRange := tEnd - tStart
	if totalRange <= 0 {
		return false
	}

	var coveredDuration time.Duration
	for _, caption := range captions {
		// Calculate overlap with the specified range
		overlapStart := maxDuration(caption.StartTime, tStart)
		overlapEnd := minDuration(caption.EndTime, tEnd)

		if overlapStart < overlapEnd {
			coveredDuration += overlapEnd - overlapStart
		}
	}

	actualCoverage := float64(coveredDuration) / float64(totalRange)
	return actualCoverage >= requiredCoverage
}

func validateLanguage(text, endpoint string) bool {
	if text == "" {
		return false
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(endpoint, "text/plain", strings.NewReader(text))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	var langResp models.LangResponse
	if err := json.Unmarshal(body, &langResp); err != nil {
		return false
	}

	return langResp.Lang == "en-US"
}



func printValidationError(errorType, description string) {
	validationError := models.ValidationError{
		Type:        errorType,
		Description: description,
	}
	jsonBytes, _ := json.Marshal(validationError)
	fmt.Println(string(jsonBytes))
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
