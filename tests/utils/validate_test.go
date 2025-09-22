package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theCompanyDream/srt-test/internal/models"
	"github.com/theCompanyDream/srt-test/internal/utils"
)

func TestIsValidFileType(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{"valid vtt file", "captions.vtt", true},
		{"valid srt file", "captions.srt", true},
		{"valid vtt uppercase", "captions.VTT", true},
		{"valid srt uppercase", "captions.SRT", true},
		{"valid with path", "/path/to/captions.vtt", true},
		{"invalid txt file", "captions.txt", false},
		{"invalid no extension", "captions", false},
		{"invalid with dot", "captions.vtt.backup", false},
		{"invalid empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsValidFileType(tt.filePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateCoverage(t *testing.T) {
	baseTime := time.Duration(0)
	oneSec := time.Second
	twoSec := 2 * time.Second
	threeSec := 3 * time.Second

	tests := []struct {
		name             string
		captions         []models.CaptionEntry
		tStart           time.Duration
		tEnd             time.Duration
		requiredCoverage float64
		expected         bool
	}{
		{
			name: "full coverage",
			captions: []models.CaptionEntry{
				{StartTime: baseTime, EndTime: oneSec},
				{StartTime: oneSec, EndTime: twoSec},
			},
			tStart:           baseTime,
			tEnd:             twoSec,
			requiredCoverage: 1.0,
			expected:         true,
		},
		{
			name: "50% coverage meets 50% requirement",
			captions: []models.CaptionEntry{
				{StartTime: baseTime, EndTime: oneSec},
			},
			tStart:           baseTime,
			tEnd:             twoSec,
			requiredCoverage: 0.5,
			expected:         true,
		},
		{
			name: "33% coverage below 50% requirement",
			captions: []models.CaptionEntry{
				{StartTime: baseTime, EndTime: oneSec},
			},
			tStart:           baseTime,
			tEnd:             threeSec,
			requiredCoverage: 0.5,
			expected:         false,
		},
		{
			name:             "no captions",
			captions:         []models.CaptionEntry{},
			tStart:           baseTime,
			tEnd:             twoSec,
			requiredCoverage: 0.5,
			expected:         false,
		},
		{
			name: "invalid time range",
			captions: []models.CaptionEntry{
				{StartTime: baseTime, EndTime: oneSec},
			},
			tStart:           twoSec,
			tEnd:             baseTime,
			requiredCoverage: 0.5,
			expected:         false,
		},
		{
			name: "zero time range",
			captions: []models.CaptionEntry{
				{StartTime: baseTime, EndTime: oneSec},
			},
			tStart:           baseTime,
			tEnd:             baseTime,
			requiredCoverage: 0.5,
			expected:         false,
		},
		{
			name: "overlapping captions",
			captions: []models.CaptionEntry{
				{StartTime: baseTime, EndTime: twoSec},
				{StartTime: oneSec, EndTime: threeSec},
			},
			tStart:           baseTime,
			tEnd:             threeSec,
			requiredCoverage: 1.0,
			expected:         true,
		},
		{
			name: "captions outside range don't count",
			captions: []models.CaptionEntry{
				{StartTime: threeSec, EndTime: 4 * time.Second},
			},
			tStart:           baseTime,
			tEnd:             twoSec,
			requiredCoverage: 0.5,
			expected:         false,
		},
		{
			name: "exact coverage threshold",
			captions: []models.CaptionEntry{
				{StartTime: baseTime, EndTime: time.Second},
			},
			tStart:           baseTime,
			tEnd:             2 * time.Second,
			requiredCoverage: 0.5,
			expected:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ValidateCoverage(tt.captions, tt.tStart, tt.tEnd, tt.requiredCoverage)
			assert.Equal(t, tt.expected, result, tt.name)
		})
	}
}

func TestValidateLanguage(t *testing.T) {
	t.Run("empty text returns false", func(t *testing.T) {
		result := utils.ValidateLanguage("", "http://example.com")
		assert.False(t, result)
	})

	// Test with a mock server for the happy path
	t.Run("successful english detection with test server", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify the request
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))

			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			assert.Equal(t, "Hello world", string(body))

			// Respond with English detection
			response := models.LangResponse{Lang: "en-US"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		result := utils.ValidateLanguage("Hello world", server.URL)
		assert.True(t, result)
	})

	t.Run("non-english detection with test server", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := models.LangResponse{Lang: "es-ES"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		result := utils.ValidateLanguage("Hola mundo", server.URL)
		assert.False(t, result)
	})

	// Test error cases without making real HTTP calls
	t.Run("invalid endpoint returns false", func(t *testing.T) {
		// This will fail to connect, testing the error path
		result := utils.ValidateLanguage("test text", "http://invalid-endpoint-that-does-not-exist:9999")
		assert.False(t, result)
	})

	t.Run("server error returns false", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		result := utils.ValidateLanguage("test text", server.URL)
		assert.False(t, result)
	})

	t.Run("invalid JSON response returns false", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		result := utils.ValidateLanguage("test text", server.URL)
		assert.False(t, result)
	})
}

func TestPrintValidationError(t *testing.T) {
	tests := []struct {
		name         string
		errorType    string
		description  string
		expectedJSON string
	}{
		{
			name:         "format error",
			errorType:    "FormatError",
			description:  "Invalid file format",
			expectedJSON: `{"type":"FormatError","description":"Invalid file format"}`,
		},
		{
			name:         "coverage error",
			errorType:    "CoverageError",
			description:  "Only 45% coverage",
			expectedJSON: `{"type":"CoverageError","description":"Only 45% coverage"}`,
		},
		{
			name:         "language error",
			errorType:    "LanguageError",
			description:  "Non-English content detected",
			expectedJSON: `{"type":"LanguageError","description":"Non-English content detected"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			utils.PrintValidationError(tt.errorType, tt.description)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := strings.TrimSpace(buf.String())

			// Verify the output matches expected JSON
			assert.Equal(t, tt.expectedJSON, output)

			// Verify it's valid JSON that can be unmarshalled
			var ve models.ValidationError
			err := json.Unmarshal([]byte(output), &ve)
			assert.NoError(t, err)
			assert.Equal(t, tt.errorType, ve.Type)
			assert.Equal(t, tt.description, ve.Description)
		})
	}
}
