package parse

import (
	"strings"

	"github.com/theCompanyDream/srt-test/internal/models"
)

func ExtractAllText(captions []models.CaptionEntry) string {
	var textParts []string
	for _, caption := range captions {
		if strings.TrimSpace(caption.Text) != "" {
			textParts = append(textParts, caption.Text)
		}
	}
	return strings.Join(textParts, " ")
}