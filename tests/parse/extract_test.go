package parse

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/theCompanyDream/srt-test/internal/models"
	"github.com/theCompanyDream/srt-test/internal/parse"
)

func TestExtractAllText(t *testing.T) {
	tests := []struct {
		name     string
		captions []models.CaptionEntry
		expected string
	}{
		{
			name:     "empty captions list",
			captions: []models.CaptionEntry{},
			expected: "",
		},
		{
			name: "single caption with text",
			captions: []models.CaptionEntry{
				{
					StartTime: 1 * time.Second,
					EndTime:   3 * time.Second,
					Text:      "Hello world",
				},
			},
			expected: "Hello world",
		},
		{
			name: "multiple captions with text",
			captions: []models.CaptionEntry{
				{
					StartTime: 1 * time.Second,
					EndTime:   3 * time.Second,
					Text:      "Hello world",
				},
				{
					StartTime: 4 * time.Second,
					EndTime:   6 * time.Second,
					Text:      "This is a test",
				},
				{
					StartTime: 7 * time.Second,
					EndTime:   9 * time.Second,
					Text:      "Goodbye world",
				},
			},
			expected: "Hello world This is a test Goodbye world",
		},
		{
			name: "captions with empty text are filtered out",
			captions: []models.CaptionEntry{
				{
					StartTime: 1 * time.Second,
					EndTime:   3 * time.Second,
					Text:      "First caption",
				},
				{
					StartTime: 4 * time.Second,
					EndTime:   6 * time.Second,
					Text:      "", // Empty text should be filtered
				},
				{
					StartTime: 7 * time.Second,
					EndTime:   9 * time.Second,
					Text:      "   ", // Whitespace only should be filtered
				},
				{
					StartTime: 10 * time.Second,
					EndTime:   12 * time.Second,
					Text:      "Last caption",
				},
			},
			expected: "First caption Last caption",
		},
		{
			name: "captions with only whitespace text are filtered out",
			captions: []models.CaptionEntry{
				{
					StartTime: 1 * time.Second,
					EndTime:   3 * time.Second,
					Text:      "   \t\n   ",
				},
				{
					StartTime: 4 * time.Second,
					EndTime:   6 * time.Second,
					Text:      "Actual text",
				},
				{
					StartTime: 7 * time.Second,
					EndTime:   9 * time.Second,
					Text:      "   ",
				},
			},
			expected: "Actual text",
		},
		{
			name: "text with leading/trailing spaces is preserved but joined properly",
			captions: []models.CaptionEntry{
				{
					StartTime: 1 * time.Second,
					EndTime:   3 * time.Second,
					Text:      "Hello",
				},
				{
					StartTime: 4 * time.Second,
					EndTime:   6 * time.Second,
					Text:      "World",
				},
			},
			expected: "Hello World", // Note: internal spaces preserved, but joined with single space
		},
		{
			name: "mixed valid and invalid entries",
			captions: []models.CaptionEntry{
				{
					StartTime: 1 * time.Second,
					EndTime:   3 * time.Second,
					Text:      "Start",
				},
				{
					StartTime: 4 * time.Second,
					EndTime:   6 * time.Second,
					Text:      "", // Filtered out
				},
				{
					StartTime: 7 * time.Second,
					EndTime:   9 * time.Second,
					Text:      "Middle",
				},
				{
					StartTime: 10 * time.Second,
					EndTime:   12 * time.Second,
					Text:      "   ", // Filtered out
				},
				{
					StartTime: 13 * time.Second,
					EndTime:   15 * time.Second,
					Text:      "End",
				},
			},
			expected: "Start Middle End",
		},
		{
			name: "captions with special characters and numbers",
			captions: []models.CaptionEntry{
				{
					StartTime: 1 * time.Second,
					EndTime:   3 * time.Second,
					Text:      "Hello 123!",
				},
				{
					StartTime: 4 * time.Second,
					EndTime:   6 * time.Second,
					Text:      "Test @#$%",
				},
				{
					StartTime: 7 * time.Second,
					EndTime:   9 * time.Second,
					Text:      "Numbers 1, 2, 3",
				},
			},
			expected: "Hello 123! Test @#$% Numbers 1, 2, 3",
		},
		{
			name: "all captions have empty text",
			captions: []models.CaptionEntry{
				{
					StartTime: 1 * time.Second,
					EndTime:   3 * time.Second,
					Text:      "",
				},
				{
					StartTime: 4 * time.Second,
					EndTime:   6 * time.Second,
					Text:      "   ",
				},
				{
					StartTime: 7 * time.Second,
					EndTime:   9 * time.Second,
					Text:      "\t\n",
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parse.ExtractAllText(tt.captions)
			assert.Equal(t, tt.expected, result)

			// Additional validation: if we expect non-empty result, verify it's not empty
			if tt.expected != "" {
				assert.NotEmpty(t, result)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestExtractAllText_EdgeCases(t *testing.T) {
	t.Run("nil captions slice", func(t *testing.T) {
		// This should handle nil gracefully
		result := parse.ExtractAllText(nil)
		assert.Equal(t, "", result)
	})

	t.Run("single space is filtered out", func(t *testing.T) {
		captions := []models.CaptionEntry{
			{
				StartTime: 1 * time.Second,
				EndTime:   3 * time.Second,
				Text:      " ",
			},
		}
		result := parse.ExtractAllText(captions)
		assert.Equal(t, "", result)
	})

	t.Run("very long text is handled correctly", func(t *testing.T) {
		longText := strings.Repeat("word ", 1000)
		captions := []models.CaptionEntry{
			{
				StartTime: 1 * time.Second,
				EndTime:   3 * time.Second,
				Text:      longText,
			},
		}
		result := parse.ExtractAllText(captions)
		assert.Equal(t, strings.TrimSpace(longText), strings.TrimSpace(result))
	})
}

// Benchmark test for performance
func BenchmarkExtractAllText(b *testing.B) {
	// Create a realistic set of captions for benchmarking
	captions := make([]models.CaptionEntry, 1000)
	for i := 0; i < 1000; i++ {
		captions[i] = models.CaptionEntry{
			StartTime: time.Duration(i) * time.Second,
			EndTime:   time.Duration(i+1) * time.Second,
			Text:      "This is caption number",
		}
		// Every 10th caption is empty to test filtering
		if i%10 == 0 {
			captions[i].Text = ""
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parse.ExtractAllText(captions)
	}
}
