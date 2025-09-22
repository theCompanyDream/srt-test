package main

import (
	"fmt"
	"os"

	"github.com/theCompanyDream/srt-test/internal/cmd"
	"github.com/theCompanyDream/srt-test/internal/models"
	"github.com/theCompanyDream/srt-test/internal/parse"
	"github.com/theCompanyDream/srt-test/internal/utils"
)

func main() {
	config, err := cmd.ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Validate file type
	if !utils.IsValidFileType(config.FilePath) {
		os.Exit(1)
	}

	// Parse caption file
	captions, err := parse.ParseCaptionFile(config.FilePath)
	if err != nil {
		utils.PrintValidationError("file_parse_error", fmt.Sprintf("Failed to parse caption file: %v", err))
		os.Exit(0)
	}

	var validationErrors []models.ValidationError

	// Validate coverage
	if !utils.ValidateCoverage(captions, config.TStart, config.TEnd, config.Coverage) {
		validationErrors = append(validationErrors, models.ValidationError{
			Type:        "insufficient_coverage",
			Description: fmt.Sprintf("Captions do not cover required %.1f%% of time range %v to %v", config.Coverage*100, config.TStart, config.TEnd),
		})
	}

	// Extract and validate language
	allText := parse.ExtractAllText(captions)
	if !utils.ValidateLanguage(allText, config.Endpoint) {
		validationErrors = append(validationErrors, models.ValidationError{
			Type:        "invalid_language",
			Description: "Caption language is not en-US or language detection failed",
		})
	}

	// Print validation errors
	for _, err := range validationErrors {
		utils.PrintValidationError(err.Type, err.Description)
	}

	os.Exit(0)
}
