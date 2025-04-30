package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type LyricsTranslator struct {
	mock.Mock
}

func (m *LyricsTranslator) TranslateLyrics(ctx context.Context, lyrics []string) ([]string, error) {
	args := m.Called(ctx, lyrics)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
