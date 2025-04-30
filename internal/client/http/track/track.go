package track

import (
	"errors"
	"strings"
)

var (
	ErrLyricsNotFound        = errors.New("track not found")
	ErrFailedTranslateLyrics = errors.New("failed translate track")
)

func FormatLyrics(lyrics string) []string {
	normalized := strings.ReplaceAll(lyrics, "\r\n", "\n")

	lines := strings.Split(normalized, "\n")

	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
