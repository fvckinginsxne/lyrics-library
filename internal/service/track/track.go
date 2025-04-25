package track

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"lyrics-library/internal/client"
	"lyrics-library/internal/domain/model"
	"lyrics-library/internal/lib/logger/sl"
	"lyrics-library/internal/storage"
)

type LyricsProvider interface {
	Lyrics(ctx context.Context, artist, title string) ([]string, error)
}

type LyricsTranslator interface {
	TranslateLyrics(ctx context.Context, lyrics []string) ([]string, error)
}

type Storage interface {
	SaveTrack(ctx context.Context, track *model.Track) error
	Track(ctx context.Context, artist, title string) (*model.Track, error)
	TracksByArtist(ctx context.Context, artist string) ([]*model.Track, error)
	DeleteTrack(ctx context.Context, uuid string) error
}

type Cache interface {
	SaveArtistTracks(ctx context.Context, artist string, tracks []*model.Track) error
	ArtistTracks(ctx context.Context, artist string) ([]*model.Track, error)
	Track(ctx context.Context, artist, title string) (*model.Track, error)
	SaveTrack(ctx context.Context, track *model.Track) error
}

var (
	ErrLyricsNotFound        = errors.New("lyrics not found")
	ErrFailedTranslateLyrics = errors.New("failed to translate lyrics")
	ErrTrackNotFound         = errors.New("track not found")
	ErrArtistTracksNotFound  = errors.New("artist's tracks not found")
	ErrInvalidUUID           = errors.New("invalid uuid")
)

type Service struct {
	log              *slog.Logger
	lyricsProvider   LyricsProvider
	lyricsTranslator LyricsTranslator
	storage          Storage
	cache            Cache
}

func New(
	log *slog.Logger,
	lyricsProvider LyricsProvider,
	lyricsTranslator LyricsTranslator,
	storage Storage,
	cache Cache,
) *Service {
	return &Service{
		log:              log,
		lyricsProvider:   lyricsProvider,
		lyricsTranslator: lyricsTranslator,
		storage:          storage,
		cache:            cache,
	}
}

func (s *Service) Save(
	ctx context.Context,
	artist, title string,
) (*model.Track, error) {
	const op = "service.track.Save"

	log := s.log.With("op", op)

	log.Info("saving track")

	cached, err := s.cache.Track(ctx, artist, title)
	if err == nil {
		log.Info("returning cached track")

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

	track := &model.Track{
		Artist:      artist,
		Title:       title,
		Lyrics:      lyrics,
		Translation: translation,
	}

	if err := s.storage.SaveTrack(ctx, track); err != nil {
		log.Error("failed to save track", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	go func() {
		log.Info("saving track in cache")

		if err := s.cache.SaveTrack(ctx, track); err != nil {
			log.Error("failed to cache track", sl.Err(err))
		}
	}()

	log.Info("lyrics saved successfully")

	return track, nil
}

func (s *Service) Track(
	ctx context.Context,
	artist, title string,
) (*model.Track, error) {
	const op = "service.track.Track"

	log := s.log.With(slog.String("op", op))

	log.Info("getting track")

	cached, err := s.cache.Track(ctx, artist, title)
	if err == nil {
		log.Info("returnig cached track")

		return cached, nil
	}

	track, err := s.storage.Track(ctx, artist, title)
	if err != nil {
		log.Error("failed to get track", sl.Err(err))

		if errors.Is(err, storage.ErrTrackNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrTrackNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	go func() {
		log.Info("caching track")

		if err := s.cache.SaveTrack(ctx, track); err != nil {
			log.Error("failed to cache track", sl.Err(err))
		}
	}()

	log.Info("track got successfully")

	return track, nil
}

func (s *Service) ArtistTracks(ctx context.Context, artist string) ([]*model.Track, error) {
	const op = "service.track.ArtistTracks"

	log := s.log.With(slog.String("op", op))

	cached, err := s.cache.ArtistTracks(ctx, artist)
	if err == nil {
		log.Info("getting tracks from cache")

		return cached, nil
	}

	tracks, err := s.storage.TracksByArtist(ctx, artist)
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

		if err := s.cache.SaveArtistTracks(ctx, artist, tracks); err != nil {
			log.Error("failed to cache artist tracks", sl.Err(err))
		}
	}()

	log.Info("artist's tracks got successfully", slog.Any("tracks", tracks))

	return tracks, nil
}

func (s *Service) Delete(ctx context.Context, uuid string) error {
	const op = "service.track.Delete"

	log := s.log.With(slog.String("op", op))

	log.Info("deleting track by uuid")

	if err := s.storage.DeleteTrack(ctx, uuid); err != nil {
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
