package read

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	trackService "lyrics-library/internal/service/track"
	"lyrics-library/internal/transport/dto"
)

type TrackProvider interface {
	Track(ctx context.Context, artist, title string) (*dto.TrackResponse, error)
}

type ArtistTracksProvider interface {
	ArtistTracks(ctx context.Context, artist string) ([]*dto.TrackResponse, error)
}

// @Summary Get song lyrics or artist tracks
// @Description If 'title' is provided, returns lyrics for the specific song.
// @Description Otherwise, returns a list of all songs by the artist (without track).
// @Tags track
// @Param artist query string true "Artist name" example("Juice WRLD")
// @Param title query string false "Song title (optional)" example("Legends")
// @Success 200 {object} dto.TrackResponse "Returns lyrics (object) or artist tracks (array)"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /track [get]
func New(
	ctx context.Context,
	log *slog.Logger,
	trackProvider TrackProvider,
	artistTracksProvider ArtistTracksProvider,
) gin.HandlerFunc {
	const op = "handler.track.read.New"

	return func(c *gin.Context) {
		log = log.With(slog.String("op", op))

		log.Info("getting track")

		artist := c.Query("artist")
		title := c.Query("title")

		if artist == "" {
			log.Error("missing 'artist' parameter")

			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "artist is required"})
			return
		}

		if title == "" {
			tracks, err := artistTracksProvider.ArtistTracks(ctx, artist)
			if err != nil {
				if errors.Is(err, trackService.ErrArtistTracksNotFound) {
					c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "artist tracks not found"})
					return
				}

				c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
				return
			}

			c.JSON(http.StatusOK, tracks)
			return
		}

		track, err := trackProvider.Track(ctx, artist, title)
		if err != nil {
			if errors.Is(err, trackService.ErrTrackNotFound) {
				c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "track not found"})
				return
			}

			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
			return
		}

		c.JSON(http.StatusOK, track)
	}
}
