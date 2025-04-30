package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"lyrics-library/internal/domain/model"
)

type Cache struct {
	mock.Mock
}

func (m *Cache) SaveArtistTracks(ctx context.Context, artist string, tracks []*model.Track) error {
	args := m.Called(ctx, artist, tracks)
	return args.Error(0)
}

func (m *Cache) ArtistTracks(ctx context.Context, artist string) ([]*model.Track, error) {
	args := m.Called(ctx, artist)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Track), args.Error(1)
}

func (m *Cache) Track(ctx context.Context, artist, title string) (*model.Track, error) {
	args := m.Called(ctx, artist, title)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Track), args.Error(1)
}

func (m *Cache) SaveTrack(ctx context.Context, track *model.Track) error {
	args := m.Called(ctx, track)
	return args.Error(0)
}
