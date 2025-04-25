package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"

	"lyrics-library/internal/domain/model"
	"lyrics-library/internal/storage"
)

type Storage struct {
	db *redis.Client
}

func New(redisURL, password string) (*Storage, error) {
	const op = "storage.redis.New"

	db := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: password,
		DB:       0,
	})

	if _, err := db.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) SaveTrack(ctx context.Context, track *model.Track) error {
	const op = "storage.redis.SaveTrack"

	key := generateTrackKey(track.Artist, track.Title)

	data, err := json.Marshal(track)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = s.db.Set(ctx, key, data, 0).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) Track(ctx context.Context, artist, title string) (*model.Track, error) {
	const op = "storage.redis.GetTrack"

	key := generateTrackKey(artist, title)

	data, err := s.db.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrTrackNotCached)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var track model.Track
	if err := json.Unmarshal(data, &track); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &track, err
}

func (s *Storage) SaveArtistTracks(ctx context.Context, artist string, tracks []*model.Track) error {
	const op = "storage.redis.SaveArtistTracks"

	key := generateArtistTracksKey(artist)

	data, err := json.Marshal(tracks)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.db.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) ArtistTracks(ctx context.Context, artist string) ([]*model.Track, error) {
	const op = "storage.redis.GetArtistTracks"

	key := generateArtistTracksKey(artist)

	data, err := s.db.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, storage.ErrArtistTracksNotCached) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrArtistTracksNotCached)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var tracks []*model.Track
	if err := json.Unmarshal(data, &tracks); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tracks, nil
}

func (s *Storage) Close(ctx context.Context) error {
	if err := s.db.Close(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.db.Ping(ctx).Err()
}

func generateArtistTracksKey(artist string) string {
	return fmt.Sprintf("artist_tracks:%s", artist)
}

func generateTrackKey(artist, title string) string {
	return fmt.Sprintf("track:%s:%s", artist, title)
}
