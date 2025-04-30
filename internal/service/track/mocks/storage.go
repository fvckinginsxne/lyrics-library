package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"lyrics-library/internal/domain/model"
)

type Storage struct {
	mock.Mock
}

func (m *Storage) SaveTrack(ctx context.Context, track *model.Track) error {
	args := m.Called(ctx, track)
	return args.Error(0)
}

func (m *Storage) Track(ctx context.Context, artist, title string) (*model.Track, error) {
	args := m.Called(ctx, artist, title)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Track), args.Error(1)
}

func (m *Storage) TracksByArtist(ctx context.Context, artist string) ([]*model.Track, error) {
	args := m.Called(ctx, artist)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Track), args.Error(1)
}

func (m *Storage) DeleteTrack(ctx context.Context, uuid string) error {
	args := m.Called(ctx, uuid)
	return args.Error(0)
}
