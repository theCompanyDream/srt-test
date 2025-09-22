package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/theCompanyDream/srt-test/internal/utils"
)

func TestMaxDuration(t *testing.T) {
	tests := []struct {
		name     string
		a        time.Duration
		b        time.Duration
		expected time.Duration
	}{
		{
			name:     "a greater than b",
			a:        10 * time.Second,
			b:        5 * time.Second,
			expected: 10 * time.Second,
		},
		{
			name:     "b greater than a",
			a:        3 * time.Second,
			b:        7 * time.Second,
			expected: 7 * time.Second,
		},
		{
			name:     "equal durations",
			a:        5 * time.Minute,
			b:        5 * time.Minute,
			expected: 5 * time.Minute,
		},
		{
			name:     "zero and positive",
			a:        0,
			b:        2 * time.Hour,
			expected: 2 * time.Hour,
		},
		{
			name:     "negative durations",
			a:        -10 * time.Second,
			b:        -5 * time.Second,
			expected: -5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.MaxDuration(tt.a, tt.b)
			assert.Equal(t, tt.expected, result, "maxDuration(%v, %v) should equal %v", tt.a, tt.b, tt.expected)
		})
	}
}

func TestMinDuration(t *testing.T) {
	tests := []struct {
		name     string
		a        time.Duration
		b        time.Duration
		expected time.Duration
	}{
		{
			name:     "a less than b",
			a:        2 * time.Second,
			b:        8 * time.Second,
			expected: 2 * time.Second,
		},
		{
			name:     "b less than a",
			a:        15 * time.Minute,
			b:        10 * time.Minute,
			expected: 10 * time.Minute,
		},
		{
			name:     "equal durations",
			a:        30 * time.Millisecond,
			b:        30 * time.Millisecond,
			expected: 30 * time.Millisecond,
		},
		{
			name:     "zero and positive",
			a:        0,
			b:        1 * time.Hour,
			expected: 0,
		},
		{
			name:     "negative durations",
			a:        -3 * time.Second,
			b:        -8 * time.Second,
			expected: -8 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.MinDuration(tt.a, tt.b)
			assert.Equal(t, tt.expected, result, "minDuration(%v, %v) should equal %v", tt.a, tt.b, tt.expected)
		})
	}
}
