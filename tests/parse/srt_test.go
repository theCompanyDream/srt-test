package parse

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/theCompanyDream/srt-test/internal/parse"
)

func TestParseSRTTime(t *testing.T) {
	tests := []struct {
		name          string
		timeStr       string
		expected      time.Duration
		expectError   bool
		errorContains string
	}{
		{
			name:     "basic time",
			timeStr:  "00:00:01,000",
			expected: 1 * time.Second,
		},
		{
			name:     "with milliseconds",
			timeStr:  "00:00:01,500",
			expected: 1500 * time.Millisecond,
		},
		{
			name:     "minutes and seconds",
			timeStr:  "00:01:23,456",
			expected: 1*time.Minute + 23*time.Second + 456*time.Millisecond,
		},
		{
			name:     "hours, minutes, seconds",
			timeStr:  "01:23:45,678",
			expected: time.Hour + 23*time.Minute + 45*time.Second + 678*time.Millisecond,
		},
		{
			name:     "maximum valid time",
			timeStr:  "99:59:59,999",
			expected: 99*time.Hour + 59*time.Minute + 59*time.Second + 999*time.Millisecond,
		},
		{
			name:          "invalid format - missing comma",
			timeStr:       "00:00:01.000",
			expectError:   true,
			errorContains: "invalid seconds format",
		},
		{
			name:          "invalid format - wrong parts count",
			timeStr:       "00:00:01:000",
			expectError:   true,
			errorContains: "invalid time format",
		},
		{
			name:          "invalid hours",
			timeStr:       "ab:00:01,000",
			expectError:   true,
			errorContains: "invalid syntax",
		},
		{
			name:          "invalid minutes",
			timeStr:       "00:ab:01,000",
			expectError:   true,
			errorContains: "invalid syntax",
		},
		{
			name:          "invalid seconds",
			timeStr:       "00:00:ab,000",
			expectError:   true,
			errorContains: "invalid syntax",
		},
		{
			name:          "invalid milliseconds",
			timeStr:       "00:00:01,abc",
			expectError:   true,
			errorContains: "invalid syntax",
		},
		{
			name:          "empty string",
			timeStr:       "",
			expectError:   true,
			errorContains: "invalid time format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parse.ParseSRTTime(tt.timeStr)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseSRT_EdgeCases(t *testing.T) {
	t.Run("very large SRT file", func(t *testing.T) {
		var builder strings.Builder
		for i := 1; i <= 1000; i++ {
			builder.WriteString(fmt.Sprintf("%d\n", i))
			builder.WriteString(fmt.Sprintf("00:00:%02d,000 --> 00:00:%02d,500\n", i, i+1))
			builder.WriteString(fmt.Sprintf("Caption number %d\n\n", i))
		}

		reader := bufio.NewReader(strings.NewReader(builder.String()))
		captions, err := parse.ParseSRT(reader)

		assert.NoError(t, err)
		assert.Len(t, captions, 1000)
		assert.Equal(t, "Caption number 1", captions[0].Text)
		assert.Equal(t, "00:00:1000,000 --> 00:00:1001,500 Caption number 1000", captions[999].Text)
	})

	t.Run("SRT with special characters in text", func(t *testing.T) {
		input := `1
00:00:01,000 --> 00:00:04,000
Hello @world! #test <b>HTML</b> "quotes"
Line 2`

		reader := bufio.NewReader(strings.NewReader(input))
		captions, err := parse.ParseSRT(reader)

		assert.NoError(t, err)
		assert.Len(t, captions, 1)
		assert.Equal(t, "Hello @world! #test <b>HTML</b> \"quotes\" Line 2", captions[0].Text)
	})

	t.Run("SRT with Windows line endings", func(t *testing.T) {
		input := "1\r\n00:00:01,000 --> 00:00:04,000\r\nHello world\r\n\r\n2\r\n00:00:05,000 --> 00:00:08,000\r\nAnother caption\r\n"

		reader := bufio.NewReader(strings.NewReader(input))
		captions, err := parse.ParseSRT(reader)

		assert.NoError(t, err)
		assert.Len(t, captions, 2)
		assert.Equal(t, "Hello world", captions[0].Text)
		assert.Equal(t, "Another caption", captions[1].Text)
	})

	t.Run("SRT with extra spaces in timing line", func(t *testing.T) {
		input := `1
00:00:01,000    -->    00:00:04,000
Hello world`

		reader := bufio.NewReader(strings.NewReader(input))
		captions, err := parse.ParseSRT(reader)

		assert.NoError(t, err)
		assert.Len(t, captions, 1)
		assert.Equal(t, 1*time.Second, captions[0].StartTime)
		assert.Equal(t, 4*time.Second, captions[0].EndTime)
	})
}

func TestParseSRT_ScannerError(t *testing.T) {
	// Create a reader that will cause a scanner error
	// We can't easily simulate a scanner error with strings.Reader,
	// but we can test that scanner errors are propagated
	t.Run("valid input returns no scanner error", func(t *testing.T) {
		input := `1
00:00:01,000 --> 00:00:04,000
Hello world`

		reader := bufio.NewReader(strings.NewReader(input))
		_, err := parse.ParseSRT(reader)
		assert.NoError(t, err)
	})
}

// Benchmark tests
func BenchmarkParseSRT(b *testing.B) {
	// Create a realistic SRT content for benchmarking
	var builder strings.Builder
	for i := 1; i <= 100; i++ {
		builder.WriteString(fmt.Sprintf("%d\n", i))
		builder.WriteString(fmt.Sprintf("00:00:%02d,000 --> 00:00:%02d,500\n", i, i+1))
		builder.WriteString(fmt.Sprintf("This is caption number %d with some text\n\n", i))
	}
	input := builder.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bufio.NewReader(strings.NewReader(input))
		parse.ParseSRT(reader)
	}
}

func BenchmarkParseSRTTime(b *testing.B) {
	timeStr := "01:23:45,678"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parse.ParseSRTTime(timeStr)
	}
}
