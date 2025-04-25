package get

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"lyrics-library/internal/domain/models"
	trackService "lyrics-library/internal/service/track"
)

type TrackProvider interface {
	Track(ctx context.Context, artist, title string) (*models.Track, error)
}

type ArtistTracksProvider interface {
	ArtistTracks(ctx context.Context, artist string) ([]*models.Track, error)
}

func New(
	ctx context.Context,
	log *slog.Logger,
	trackProvider TrackProvider,
	artistTracksProvider ArtistTracksProvider,
) gin.HandlerFunc {
	const op = "handlers.song.read.New"

	return func(c *gin.Context) {
		log = log.With(slog.String("op", op))

		log.Info("getting lyrics")

		artist := c.Query("artist")
		title := c.Query("title")

		if artist == "" {
			log.Error("missing 'artist' parameter")

			c.JSON(http.StatusBadRequest, gin.H{"error": "artist is required"})
			return
		}

		if title == "" {
			tracks, err := artistTracksProvider.ArtistTracks(ctx, artist)
			if err != nil {
				if errors.Is(err, trackService.ErrArtistTracksNotFound) {
					c.JSON(http.StatusNotFound, gin.H{"error": "artist tracks not found"})
					return
				}

				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}

			c.JSON(http.StatusOK, tracks)
			return
		}

		track, err := trackProvider.Track(ctx, artist, title)
		if err != nil {
			if errors.Is(err, trackService.ErrTrackNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "track not found"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, track)
	}
}
