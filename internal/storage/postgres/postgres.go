package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"lyrics-library/internal/domain/model"
	"lyrics-library/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(dbURL string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveTrack(ctx context.Context, track *model.Track) error {
	const op = "storage.postgres.Save"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO songs (artist, title, lyrics, translation)
		VALUES ($1, $2, $3, $4)
	`, track.Artist, track.Title, pq.Array(track.Lyrics), pq.Array(track.Translation))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return tx.Commit()
}

func (s *Storage) Track(ctx context.Context, artist, title string) (*model.Track, error) {
	const op = "storage.postgres.TrackInfo"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, `
		SELECT artist, title, lyrics, translation FROM songs 
		WHERE artist ILIKE $1 AND title ILIKE $2
	`, artist, title)

	var lyrics, translation []string

	err = row.Scan(&artist, &title, pq.Array(&lyrics), pq.Array(&translation))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrTrackNotFound
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &model.Track{
		Artist:      artist,
		Title:       title,
		Lyrics:      lyrics,
		Translation: translation,
	}, nil
}

func (s *Storage) TracksByArtist(ctx context.Context, artist string) ([]*model.Track, error) {
	const op = "storage.postgres.TracksByArtist"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT artist, title, lyrics, translation
		FROM songs WHERE artist ILIKE $1
	`, artist)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var tracks []*model.Track

	var (
		title       string
		lyrics      []string
		translation []string
	)
	for rows.Next() {
		err := rows.Scan(&artist, &title, pq.Array(&lyrics), pq.Array(&translation))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		tracks = append(tracks, &model.Track{
			Artist:      artist,
			Title:       title,
			Lyrics:      lyrics,
			Translation: translation,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(tracks) == 0 {
		return nil, fmt.Errorf("%s: %w", op, storage.ErrArtistTracksNotFound)
	}

	return tracks, nil
}

func (s *Storage) DeleteTrack(ctx context.Context, uuid string) error {
	const op = "storage.postgres.DeleteTrack"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `DELETE FROM songs WHERE uuid = $1`, uuid)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrInvalidUUID)
	}

	return tx.Commit()
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Storage) Close(ctx context.Context) error {
	done := make(chan struct{})

	var closeErr error
	go func() {
		closeErr = s.db.Close()
		close(done)
	}()

	select {
	case <-done:
		return closeErr
	case <-ctx.Done():
		return ctx.Err()
	}
}
