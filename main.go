package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ValidationError represents a validation failure
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
	FilePath   string
	TStart     time.Duration
	TEnd       time.Duration
	Coverage   float64
	Endpoint   string
}

func main() {
	config, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Validate file type
	if !isValidFileType(config.FilePath) {
		os.Exit(1)
	}

	// Parse caption file
	captions, err := parseCaptionFile(config.FilePath)
	if err != nil {
		printValidationError("file_parse_error", fmt.Sprintf("Failed to parse caption file: %v", err))
		os.Exit(0)
	}

	var validationErrors []ValidationError

	// Validate coverage
	if !validateCoverage(captions, config.TStart, config.TEnd, config.Coverage) {
		validationErrors = append(validationErrors, ValidationError{
			Type:        "insufficient_coverage",
			Description: fmt.Sprintf("Captions do not cover required %.1f%% of time range %v to %v", config.Coverage*100, config.TStart, config.TEnd),
		})
	}

	// Extract and validate language
	allText := extractAllText(captions)
	if !validateLanguage(allText, config.Endpoint) {
		validationErrors = append(validationErrors, ValidationError{
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

func parseFlags() (*Config, error) {
	var (
		filePath = flag.String("file", "", "Path to caption file (required)")
		tStart   = flag.String("t_start", "0s", "Start time (e.g., 30s, 1m30s)")
		tEnd     = flag.String("t_end", "", "End time (required)")
		coverage = flag.Float64("coverage", 0.8, "Required coverage percentage (0.0-1.0)")
		endpoint = flag.String("endpoint", "", "Language detection endpoint URL (required)")
	)
	flag.Parse()

	if *filePath == "" {
		return nil, fmt.Errorf("file path is required")
	}
	if *tEnd == "" {
		return nil, fmt.Errorf("end time is required")
	}
	if *endpoint == "" {
		return nil, fmt.Errorf("endpoint URL is required")
	}

	startTime, err := time.ParseDuration(*tStart)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format: %v", err)
	}

	endTime, err := time.ParseDuration(*tEnd)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format: %v", err)
	}

	if startTime >= endTime {
		return nil, fmt.Errorf("start time must be less than end time")
	}

	if *coverage < 0 || *coverage > 1 {
		return nil, fmt.Errorf("coverage must be between 0.0 and 1.0")
	}

	return &Config{
		FilePath: *filePath,
		TStart:   startTime,
		TEnd:     endTime,
		Coverage: *coverage,
		Endpoint: *endpoint,
	}, nil
}

func isValidFileType(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".vtt" || ext == ".srt"
}

func parseCaptionFile(filePath string) ([]CaptionEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".vtt":
		return parseWebVTT(file)
	case ".srt":
		return parseSRT(file)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

func parseWebVTT(reader io.Reader) ([]CaptionEntry, error) {
	scanner := bufio.NewScanner(reader)
	var captions []CaptionEntry
	var currentEntry CaptionEntry
	var textLines []string
	inHeader := true

	timeRegex := regexp.MustCompile(`(\d{2}:\d{2}:\d{2}\.\d{3})\s+-->\s+(\d{2}:\d{2}:\d{2}\.\d{3})`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip header
		if inHeader {
			if line == "" || strings.HasPrefix(line, "WEBVTT") || strings.HasPrefix(line, "NOTE") {
				continue
			}
			inHeader = false
		}

		// Empty line indicates end of caption block
		if line == "" {
			if len(textLines) > 0 {
				currentEntry.Text = strings.Join(textLines, " ")
				captions = append(captions, currentEntry)
				textLines = nil
			}
			continue
		}

		// Check if line contains timing
		if matches := timeRegex.FindStringSubmatch(line); len(matches) == 3 {
			var err error
			currentEntry.StartTime, err = parseWebVTTTime(matches[1])
			if err != nil {
				return nil, fmt.Errorf("error parsing start time: %v", err)
			}
			currentEntry.EndTime, err = parseWebVTTTime(matches[2])
			if err != nil {
				return nil, fmt.Errorf("error parsing end time: %v", err)
			}
		} else {
			// This is text content
			textLines = append(textLines, line)
		}
	}

	// Handle last caption if file doesn't end with empty line
	if len(textLines) > 0 {
		currentEntry.Text = strings.Join(textLines, " ")
		captions = append(captions, currentEntry)
	}

	return captions, scanner.Err()
}

func parseSRT(reader io.Reader) ([]CaptionEntry, error) {
	scanner := bufio.NewScanner(reader)
	var captions []CaptionEntry
	var currentEntry CaptionEntry
	var textLines []string
	expectingSequence := true

	timeRegex := regexp.MustCompile(`(\d{2}:\d{2}:\d{2},\d{3})\s+-->\s+(\d{2}:\d{2}:\d{2},\d{3})`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Empty line indicates end of caption block
		if line == "" {
			if len(textLines) > 0 {
				currentEntry.Text = strings.Join(textLines, " ")
				captions = append(captions, currentEntry)
				textLines = nil
			}
			expectingSequence = true
			continue
		}

		// Skip sequence number
		if expectingSequence {
			expectingSequence = false
			continue
		}

		// Check if line contains timing
		if matches := timeRegex.FindStringSubmatch(line); len(matches) == 3 {
			var err error
			currentEntry.StartTime, err = parseSRTTime(matches[1])
			if err != nil {
				return nil, fmt.Errorf("error parsing start time: %v", err)
			}
			currentEntry.EndTime, err = parseSRTTime(matches[2])
			if err != nil {
				return nil, fmt.Errorf("error parsing end time: %v", err)
			}
		} else {
			// This is text content
			textLines = append(textLines, line)
		}
	}

	// Handle last caption if file doesn't end with empty line
	if len(textLines) > 0 {
		currentEntry.Text = strings.Join(textLines, " ")
		captions = append(captions, currentEntry)
	}

	return captions, scanner.Err()
}

func parseWebVTTTime(timeStr string) (time.Duration, error) {
	// Format: HH:MM:SS.mmm
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	secParts := strings.Split(parts[2], ".")
	if len(secParts) != 2 {
		return 0, fmt.Errorf("invalid seconds format: %s", parts[2])
	}

	seconds, err := strconv.Atoi(secParts[0])
	if err != nil {
		return 0, err
	}

	milliseconds, err := strconv.Atoi(secParts[1])
	if err != nil {
		return 0, err
	}

	totalMilliseconds := hours*3600000 + minutes*60000 + seconds*1000 + milliseconds
	return time.Duration(totalMilliseconds) * time.Millisecond, nil
}

func parseSRTTime(timeStr string) (time.Duration, error) {
	// Format: HH:MM:SS,mmm
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	secParts := strings.Split(parts[2], ",")
	if len(secParts) != 2 {
		return 0, fmt.Errorf("invalid seconds format: %s", parts[2])
	}

	seconds, err := strconv.Atoi(secParts[0])
	if err != nil {
		return 0, err
	}

	milliseconds, err := strconv.Atoi(secParts[1])
	if err != nil {
		return 0, err
	}

	totalMilliseconds := hours*3600000 + minutes*60000 + seconds*1000 + milliseconds
	return time.Duration(totalMilliseconds) * time.Millisecond, nil
}

func validateCoverage(captions []CaptionEntry, tStart, tEnd time.Duration, requiredCoverage float64) bool {
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

	var langResp LangResponse
	if err := json.Unmarshal(body, &langResp); err != nil {
		return false
	}

	return langResp.Lang == "en-US"
}

func extractAllText(captions []CaptionEntry) string {
	var textParts []string
	for _, caption := range captions {
		if strings.TrimSpace(caption.Text) != "" {
			textParts = append(textParts, caption.Text)
		}
	}
	return strings.Join(textParts, " ")
}

func printValidationError(errorType, description string) {
	validationError := ValidationError{
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