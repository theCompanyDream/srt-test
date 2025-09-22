package parse

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/theCompanyDream/srt-test/internal/models"
)

func ParseWebVTT(reader io.Reader) ([]models.CaptionEntry, error) {
	scanner := bufio.NewScanner(reader)
	var captions []models.CaptionEntry
	var currentEntry models.CaptionEntry
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
