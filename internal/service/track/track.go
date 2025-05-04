package track

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	trackClient "lyrics-library/internal/client/http/track"
	"lyrics-library/internal/domain/model"
	"lyrics-library/internal/lib/logger/sl"
	"lyrics-library/internal/storage"
	"lyrics-library/internal/transport/dto"
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
	ErrLyricsNotFound        = errors.New("track not found")
	ErrFailedTranslateLyrics = errors.New("failed to translate track")
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
) (*dto.TrackResponse, error) {
	const op = "service.track.Save"

	log := s.log.With("op", op)

	log.Info("saving track")

	cached, err := s.cache.Track(ctx, artist, title)
	if err == nil {
		log.Info("returning cached track")

		return dto.ToTrackResponse(cached), nil
	}

	lyrics, err := s.lyricsProvider.Lyrics(ctx, artist, title)
	if err != nil {
		if errors.Is(err, trackClient.ErrLyricsNotFound) {
			log.Error("track not found", sl.Err(err))

			return nil, fmt.Errorf("%s: %w", op, ErrLyricsNotFound)
		}

		log.Error("failed to fetch track", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("track fetched", slog.Any("track", lyrics))

	translation, err := s.lyricsTranslator.TranslateLyrics(ctx, lyrics)
	if err != nil {
		log.Error("failed translate track", sl.Err(err))

		if errors.Is(err, trackClient.ErrFailedTranslateLyrics) {

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
		log.Error("failed to create track", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	go func() {
		log.Info("saving track in cache")

		if err := s.cache.SaveTrack(ctx, track); err != nil {
			log.Error("failed to cache track", sl.Err(err))
		}
	}()

	log.Info("track saved successfully")

	return dto.ToTrackResponse(track), nil
}

func (s *Service) Track(
	ctx context.Context,
	artist, title string,
) (*dto.TrackResponse, error) {
	const op = "service.track.Track"

	log := s.log.With(slog.String("op", op))

	log.Info("getting track")

	cached, err := s.cache.Track(ctx, artist, title)
	if err == nil {
		log.Info("returning cached track")

		return dto.ToTrackResponse(cached), nil
	}

	track, err := s.storage.Track(ctx, artist, title)
	if err != nil {
		log.Error("failed to read track", sl.Err(err))

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

	return dto.ToTrackResponse(track), nil
}

func (s *Service) ArtistTracks(ctx context.Context, artist string) ([]*dto.TrackResponse, error) {
	const op = "service.track.ArtistTracks"

	log := s.log.With(slog.String("op", op))

	cached, err := s.cache.ArtistTracks(ctx, artist)
	if err == nil {
		log.Info("getting tracks from cache")

		return dto.TracksToTrackResponses(cached), nil
	}

	tracks, err := s.storage.TracksByArtist(ctx, artist)
	if err != nil {
		if errors.Is(err, storage.ErrArtistTracksNotFound) {
			log.Error("artist's track not found")

			return nil, fmt.Errorf("%s: %w", op, ErrArtistTracksNotFound)
		}

		log.Error("failed to read tracks by artist", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	go func() {
		log.Info("caching artist's tracks")

		if err := s.cache.SaveArtistTracks(ctx, artist, tracks); err != nil {
			log.Error("failed to cache artist tracks", sl.Err(err))
		}
	}()

	log.Info("artist's tracks got successfully", slog.Any("tracks", tracks))

	return dto.TracksToTrackResponses(tracks), nil
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
