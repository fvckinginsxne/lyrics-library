package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type LyricsProvider struct {
	mock.Mock
}

func (m *LyricsProvider) Lyrics(ctx context.Context, artist, title string) ([]string, error) {
	args := m.Called(ctx, artist, title)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
