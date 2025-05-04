package track

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"lyrics-library/internal/domain/model"
	"lyrics-library/internal/service/track/mocks"
)

type Mocks struct {
	lyricsProvider   *mocks.LyricsProvider
	lyricsTranslator *mocks.LyricsTranslator
	storage          *mocks.Storage
	cache            *mocks.Cache
}

func setupService(t *testing.T) (*Service, *Mocks) {
	m := &Mocks{
		lyricsProvider:   new(mocks.LyricsProvider),
		lyricsTranslator: new(mocks.LyricsTranslator),
		storage:          new(mocks.Storage),
		cache:            new(mocks.Cache),
	}

	t.Cleanup(func() {
		m.lyricsProvider.AssertExpectations(t)
		m.lyricsTranslator.AssertExpectations(t)
		m.storage.AssertExpectations(t)
		m.cache.AssertExpectations(t)
	})

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	s := New(log, m.lyricsProvider, m.lyricsTranslator, m.storage, m.cache)

	return s, m
}

func TestService_Save(t *testing.T) {
	tests := []struct {
		name          string
		artist        string
		title         string
		mockSetup     func(*Mocks)
		expectedTrack *model.Track
		expectedError error
	}{
		{
			name:   "successful create with cache hit",
			artist: "Artist1",
			title:  "Song1",
			mockSetup: func(m *Mocks) {
				m.cache.On("Track", mock.Anything, "Artist1", "Song1").
					Return(&model.Track{
						Artist: "Artist1",
						Title:  "Song1",
					}, nil)
			},
			expectedTrack: &model.Track{
				Artist: "Artist1",
				Title:  "Song1",
			},
		},
		{
			name:   "successful create with new track",
			artist: "Artist2",
			title:  "Song2",
			mockSetup: func(m *Mocks) {
				m.cache.On("Track", mock.Anything, "Artist2", "Song2").
					Return(nil, errors.New("not found"))
				m.lyricsProvider.On("Lyrics", mock.Anything, "Artist2", "Song2").
					Return([]string{"track"}, nil)
				m.lyricsTranslator.On("TranslateLyrics", mock.Anything, []string{"track"}).
					Return([]string{"translation"}, nil)
				m.storage.On("SaveTrack", mock.Anything, mock.AnythingOfType("*model.Track")).
					Return(nil)
			},
			expectedTrack: &model.Track{
				Artist:      "Artist2",
				Title:       "Song2",
				Lyrics:      []string{"track"},
				Translation: []string{"translation"},
			},
		},
		{
			name:   "track not found",
			artist: "Unknown",
			title:  "Song",
			mockSetup: func(m *Mocks) {
				m.cache.On("Track", mock.Anything, "Unknown", "Song").
					Return(nil, errors.New("not found"))
				m.lyricsProvider.On("Lyrics", mock.Anything, "Unknown", "Song").
					Return(nil, ErrLyricsNotFound)
			},
			expectedError: ErrLyricsNotFound,
		},
		{
			name:   "translation failed",
			artist: "Artist",
			title:  "Song",
			mockSetup: func(m *Mocks) {
				m.cache.On("Track", mock.Anything, "Artist", "Song").
					Return(nil, errors.New("not found"))
				m.lyricsProvider.On("Lyrics", mock.Anything, "Artist", "Song").
					Return([]string{"track"}, nil)
				m.lyricsTranslator.On("TranslateLyrics", mock.Anything, []string{"track"}).
					Return(nil, ErrFailedTranslateLyrics)
			},
			expectedError: ErrFailedTranslateLyrics,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, m := setupService(t)
			tt.mockSetup(m)

			track, err := s.Save(context.Background(), tt.artist, tt.title)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, track)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTrack.Artist, track.Artist)
				assert.Equal(t, tt.expectedTrack.Title, track.Title)

				if tt.expectedTrack.Lyrics != nil {
					assert.Equal(t, tt.expectedTrack.Lyrics, track.Lyrics)
				}

				if tt.expectedTrack.Translation != nil {
					assert.Equal(t, tt.expectedTrack.Translation, track.Translation)
				}
			}

			m.lyricsProvider.AssertExpectations(t)
			m.lyricsTranslator.AssertExpectations(t)
			m.storage.AssertExpectations(t)
			m.cache.AssertExpectations(t)
		})
	}
}

func TestService_Track(t *testing.T) {
	tests := []struct {
		name          string
		artist        string
		title         string
		mockSetup     func(*Mocks)
		expectedTrack *model.Track
		expectedError error
	}{
		{
			name:   "cache hit",
			artist: "Artist1",
			title:  "Song1",
			mockSetup: func(m *Mocks) {
				m.cache.On("Track", mock.Anything, "Artist1", "Song1").
					Return(&model.Track{
						Artist: "Artist1",
						Title:  "Song1",
					}, nil)
			},
			expectedTrack: &model.Track{
				Artist: "Artist1",
				Title:  "Song1",
			},
		},
		{
			name:   "storage hit",
			artist: "Artist2",
			title:  "Song2",
			mockSetup: func(m *Mocks) {
				m.cache.On("Track", mock.Anything, "Artist2", "Song2").
					Return(nil, errors.New("not found"))
				m.storage.On("Track", mock.Anything, "Artist2", "Song2").
					Return(&model.Track{
						Artist: "Artist2",
						Title:  "Song2",
					}, nil)
			},
			expectedTrack: &model.Track{
				Artist: "Artist2",
				Title:  "Song2",
			},
		},
		{
			name:   "track not found",
			artist: "Unknown",
			title:  "Song",
			mockSetup: func(m *Mocks) {
				m.cache.On("Track", mock.Anything, "Unknown", "Song").
					Return(nil, errors.New("not found"))
				m.storage.On("Track", mock.Anything, "Unknown", "Song").
					Return(nil, ErrTrackNotFound)
			},
			expectedError: ErrTrackNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, m := setupService(t)
			tt.mockSetup(m)

			track, err := s.Track(context.Background(), tt.artist, tt.title)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, track)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTrack.Artist, track.Artist)
				assert.Equal(t, tt.expectedTrack.Title, track.Title)
			}

			m.storage.AssertExpectations(t)
			m.cache.AssertExpectations(t)
		})
	}
}

func TestService_ArtistTracks(t *testing.T) {
	tests := []struct {
		name           string
		artist         string
		mockSetup      func(*Mocks)
		expectedTracks []*model.Track
		expectedError  error
	}{
		{
			name:   "cache hit",
			artist: "Artist1",
			mockSetup: func(m *Mocks) {
				m.cache.On("ArtistTracks", mock.Anything, "Artist1").
					Return([]*model.Track{
						{Artist: "Artist1", Title: "Song1"},
					}, nil)
			},
			expectedTracks: []*model.Track{
				{Artist: "Artist1", Title: "Song1"},
			},
		},
		{
			name:   "storage hit",
			artist: "Artist2",
			mockSetup: func(m *Mocks) {
				m.cache.On("ArtistTracks", mock.Anything, "Artist2").
					Return(nil, errors.New("not found"))
				m.storage.On("TracksByArtist", mock.Anything, "Artist2").
					Return([]*model.Track{
						{Artist: "Artist2", Title: "Song1"},
						{Artist: "Artist2", Title: "Song2"},
					}, nil)
			},
			expectedTracks: []*model.Track{
				{Artist: "Artist2", Title: "Song1"},
				{Artist: "Artist2", Title: "Song2"},
			},
		},
		{
			name:   "artist not found",
			artist: "Unknown",
			mockSetup: func(m *Mocks) {
				m.cache.On("ArtistTracks", mock.Anything, "Unknown").
					Return(nil, errors.New("not found"))
				m.storage.On("TracksByArtist", mock.Anything, "Unknown").
					Return(nil, ErrArtistTracksNotFound)
			},
			expectedError: ErrArtistTracksNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, m := setupService(t)
			tt.mockSetup(m)

			tracks, err := s.ArtistTracks(context.Background(), tt.artist)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, tracks)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedTracks), len(tracks))
				for i := range tracks {
					assert.Equal(t, tt.expectedTracks[i].Artist, tracks[i].Artist)
					assert.Equal(t, tt.expectedTracks[i].Title, tracks[i].Title)
				}
			}

			m.storage.AssertExpectations(t)
			m.cache.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		uuid          string
		mockSetup     func(*Mocks)
		expectedError error
	}{
		{
			name: "successful delete",
			uuid: "valid-uuid",
			mockSetup: func(m *Mocks) {
				m.storage.On("DeleteTrack", mock.Anything, "valid-uuid").
					Return(nil)
			},
		},
		{
			name: "invalid uuid",
			uuid: "invalid-uuid",
			mockSetup: func(m *Mocks) {
				m.storage.On("DeleteTrack", mock.Anything, "invalid-uuid").
					Return(ErrInvalidUUID)
			},
			expectedError: ErrInvalidUUID,
		},
		{
			name: "storage error",
			uuid: "valid-uuid",
			mockSetup: func(m *Mocks) {
				m.storage.On("DeleteTrack", mock.Anything, "valid-uuid").
					Return(errors.New("storage error"))
			},
			expectedError: errors.New("storage error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, m := setupService(t)
			tt.mockSetup(m)

			err := s.Delete(context.Background(), tt.uuid)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if errors.Is(tt.expectedError, ErrInvalidUUID) {
					assert.ErrorIs(t, err, ErrInvalidUUID)
				}
			} else {
				assert.NoError(t, err)
			}

			m.storage.AssertExpectations(t)
		})
	}
}
