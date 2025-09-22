package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/theCompanyDream/srt-test/internal/models"
)

func IsValidFileType(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".vtt" || ext == ".srt"
}

func ValidateCoverage(captions []models.CaptionEntry, tStart, tEnd time.Duration, requiredCoverage float64) bool {
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

func ValidateLanguage(text, endpoint string) bool {
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

func PrintValidationError(errorType, description string) {
	validationError := models.ValidationError{
		Type:        errorType,
		Description: description,
	}
	jsonBytes, _ := json.Marshal(validationError)
	fmt.Println(string(jsonBytes))
}
