package parse

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/theCompanyDream/srt-test.git/internal/models"
)

func ParseCaptionFile(filePath string) ([]models.CaptionEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".vtt":
		return ParseWebVTT(file)
	case ".srt":
		return ParseSRT(file)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}
