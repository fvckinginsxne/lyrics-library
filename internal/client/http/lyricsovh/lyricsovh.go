package lyricsovh

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	apiClient "lyrics-library/internal/client"
)

const (
	apiBaseURL = "https://api.lyrics.ovh/v1"
)

type Client struct {
	log    *slog.Logger
	client *http.Client
}

func New(log *slog.Logger) *Client {
	return &Client{
		log:    log,
		client: &http.Client{},
	}
}

type LyricsResponse struct {
	Lyrics string `json:"lyrics"`
	Error  string `json:"error"`
}

func (c *Client) Lyrics(ctx context.Context, artist, title string) ([]string, error) {
	const op = "service.api.lyricsovh.Lyrics"

	log := c.log.With(slog.String("op", op),
		slog.String("artist", artist),
		slog.String("title", title),
	)

	log.Info("Fetching lyrics")

	ctx, cancel := context.WithTimeout(ctx, apiClient.RequestTimeout)
	defer cancel()

	apiURL, err := url.JoinPath(apiBaseURL, artist, title)
	if err != nil {
		return nil, err
	}

	log.Debug("Api URL", slog.String("url", apiURL))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	result, err := c.doAPIRequest(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("Lyrics response", slog.Any("response", result))

	formatted := apiClient.FormatLyrics(result.Lyrics)

	log.Info("lyrics fetched successfully")

	return formatted, nil
}

func (c *Client) doAPIRequest(req *http.Request) (*LyricsResponse, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result LyricsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Lyrics == "" {
		return nil, apiClient.ErrLyricsNotFound
	}

	if result.Error != "" {
		return nil, fmt.Errorf("%s", result.Error)
	}

	return &result, nil
}
