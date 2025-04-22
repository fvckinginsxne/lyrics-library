package yandex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	apiClient "lyrics-library/internal/client"
)

type Response struct {
	Translations []struct {
		Text string `json:"text"`
	} `json:"translations"`
}

type Client struct {
	log    *slog.Logger
	client *http.Client
	apiKey string
}

const (
	yandexTranslateURL = "https://translate.api.cloud.yandex.net/translate/v2/translate"
	targetLanguage     = "ru"
)

func New(log *slog.Logger, apiKey string) *Client {
	return &Client{
		log:    log,
		client: &http.Client{},
		apiKey: apiKey,
	}
}

func (c *Client) TranslateLyrics(ctx context.Context, lyrics []string) ([]string, error) {
	const op = "service.api.yandex.TranslateLyrics"

	log := c.log.With(slog.String("op", op))

	log.Info("translating lyrics")

	ctx, cancel := context.WithTimeout(ctx, apiClient.RequestTimeout)
	defer cancel()

	req, err := c.buildAPIRequest(ctx, lyrics)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	res, err := c.doAPIRequest(log, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("yandex translator response", slog.Any("response", res))

	if len(res.Translations) == 0 {
		return nil, apiClient.ErrFailedTranslateLyrics
	}

	formatted := apiClient.FormatLyrics(res.Translations[0].Text)

	log.Info("lyrics translated successfully")

	return formatted, nil
}

func (c *Client) buildAPIRequest(ctx context.Context, lyrics []string) (*http.Request, error) {
	requestData := map[string]interface{}{
		"texts":              []string{strings.Join(lyrics, "\n")},
		"targetLanguageCode": targetLanguage,
	}

	reqBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		yandexTranslateURL,
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Key "+c.apiKey)

	return req, nil
}

func (c *Client) doAPIRequest(log *slog.Logger, req *http.Request) (*Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Debug("response status", slog.String("status", resp.Status))

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res Response
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
