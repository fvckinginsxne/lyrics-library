package track

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"lyrics-library/internal/client"
	"lyrics-library/internal/domain/models"
	"lyrics-library/internal/lib/logger/sl"
	"lyrics-library/internal/storage"
)

type LyricsProvider interface {
	Lyrics(ctx context.Context, artist, title string) ([]string, error)
}

type LyricsTranslator interface {
	TranslateLyrics(ctx context.Context, lyrics []string) ([]string, error)
}

type TrackManager interface {
	SaveTrack(ctx context.Context, track *models.Track) error
	Track(ctx context.Context, artist, title string) (*models.Track, error)
	TracksByArtist(ctx context.Context, artist string) ([]*models.Track, error)
	DeleteTrack(ctx context.Context, uuid string) error
}

type TrackCache interface {
	SaveArtistTracks(ctx context.Context, artist string, tracks []*models.Track) error
	ArtistTracks(ctx context.Context, artist string) ([]*models.Track, error)
	Track(ctx context.Context, artist, title string) (*models.Track, error)
	SaveTrack(ctx context.Context, track *models.Track) error
}

var (
	ErrLyricsNotFound        = errors.New("lyrics not found")
	ErrFailedTranslateLyrics = errors.New("failed to translate lyrics")
	ErrTrackNotFound         = errors.New("track not found")
	ErrArtistTracksNotFound  = errors.New("artist's tracks not found")
	ErrInvalidUUID           = errors.New("invalid uuid")
)

type TrackService struct {
	log              *slog.Logger
	lyricsProvider   LyricsProvider
	lyricsTranslator LyricsTranslator
	trackManager     TrackManager
	trackCache       TrackCache
}

func New(
	log *slog.Logger,
	lyricsProvider LyricsProvider,
	lyricsTranslator LyricsTranslator,
	trackManager TrackManager,
	trackCache TrackCache,
) *TrackService {
	return &TrackService{
		log:              log,
		lyricsProvider:   lyricsProvider,
		lyricsTranslator: lyricsTranslator,
		trackManager:     trackManager,
		trackCache:       trackCache,
	}
}

func (s *TrackService) Save(
	ctx context.Context,
	artist, title string,
) (*models.Track, error) {
	const op = "service.track.Save"

	log := s.log.With("op", op)

	log.Info("saving track")

	cached, err := s.trackCache.Track(ctx, artist, title)
	if err == nil {
		log.Info("returnig cached track")

		return cached, nil
	}

	lyrics, err := s.lyricsProvider.Lyrics(ctx, artist, title)
	if err != nil {
		if errors.Is(err, client.ErrLyricsNotFound) {
			log.Error("lyrics not found", sl.Err(err))

			return nil, fmt.Errorf("%s: %w", op, ErrLyricsNotFound)
		}

		log.Error("failed to fetch lyrics", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("lyrics fetched", slog.Any("lyrics", lyrics))

	translation, err := s.lyricsTranslator.TranslateLyrics(ctx, lyrics)
	if err != nil {
		log.Error("failed translate lyrics", sl.Err(err))

		if errors.Is(err, client.ErrFailedTranslateLyrics) {

			return nil, fmt.Errorf("%s: %w", op, ErrFailedTranslateLyrics)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	track := &models.Track{
		Artist:      artist,
		Title:       title,
		Lyrics:      lyrics,
		Translation: translation,
	}

	if err := s.trackManager.SaveTrack(ctx, track); err != nil {
		log.Error("failed to save track", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	go func() {
		log.Info("saving track in cache")

		if err := s.trackCache.SaveTrack(ctx, track); err != nil {
			log.Error("failed to cache track", sl.Err(err))
		}
	}()

	log.Info("lyrics saved successfully")

	return track, nil
}

func (s *TrackService) Track(
	ctx context.Context,
	artist, title string,
) (*models.Track, error) {
	const op = "service.track.Track"

	log := s.log.With(slog.String("op", op))

	log.Info("getting track")

	cached, err := s.trackCache.Track(ctx, artist, title)
	if err == nil {
		log.Info("returnig cached track")

		return cached, nil
	}

	track, err := s.trackManager.Track(ctx, artist, title)
	if err != nil {
		log.Error("failed to get track", sl.Err(err))

		if errors.Is(err, storage.ErrTrackNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrTrackNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	go func() {
		log.Info("caching track")

		if err := s.trackCache.SaveTrack(ctx, track); err != nil {
			log.Error("failed to cache track", sl.Err(err))
		}
	}()

	log.Info("track got successfully")

	return track, nil
}

func (s *TrackService) ArtistTracks(ctx context.Context, artist string) ([]*models.Track, error) {
	const op = "service.track.ArtistTracks"

	log := s.log.With(slog.String("op", op))

	cached, err := s.trackCache.ArtistTracks(ctx, artist)
	if err == nil {
		log.Info("getting tracks from cache")

		return cached, nil
	}

	tracks, err := s.trackManager.TracksByArtist(ctx, artist)
	if err != nil {
		if errors.Is(err, storage.ErrArtistTracksNotFound) {
			log.Error("artist's track not found")

			return nil, fmt.Errorf("%s: %w", op, ErrArtistTracksNotFound)
		}

		log.Error("failed to get tracks by artist", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	go func() {
		log.Info("caching artist's tracks")

		if err := s.trackCache.SaveArtistTracks(ctx, artist, tracks); err != nil {
			log.Error("failed to cache artist tracks", sl.Err(err))
		}
	}()

	log.Info("artist's tracks got successfully", slog.Any("tracks", tracks))

	return tracks, nil
}

func (s *TrackService) Delete(ctx context.Context, uuid string) error {
	const op = "service.track.Delete"

	log := s.log.With(slog.String("op", op))

	log.Info("deleting track by uuid")

	if err := s.trackManager.DeleteTrack(ctx, uuid); err != nil {
		if errors.Is(err, storage.ErrInvalidUUID) {
			log.Error("invalid uuid")

			return fmt.Errorf("%s: %w", op, ErrInvalidUUID)
		}

		log.Error("failed to delete track", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("track deleted successfully")

	return nil
}
