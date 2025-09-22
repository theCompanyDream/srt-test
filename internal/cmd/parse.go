package cmd

import (
	"flag"
	"fmt"
	"time"

	"github.com/theCompanyDream/srt-test/internal/models"
)

func ParseFlags() (*models.Config, error) {
	var (
		filePath = flag.String("file", "", "Path to caption file (required)")
		tStart   = flag.String("start", "0s", "Start time (e.g., 30s, 1m30s)")
		tEnd     = flag.String("end", "", "End time (required)")
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

	return &models.Config{
		FilePath: *filePath,
		TStart:   startTime,
		TEnd:     endTime,
		Coverage: *coverage,
		Endpoint: *endpoint,
	}, nil
}
