package api

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrTrackNotFound         = errors.New("track not found")
	ErrFailedTranslateLyrics = errors.New("failed translate lyrics")
)

const (
	RequestTimeout = 10 * time.Second
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
