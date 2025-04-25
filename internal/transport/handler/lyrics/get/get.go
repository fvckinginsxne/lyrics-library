package get

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"lyrics-library/internal/domain/model"
	trackService "lyrics-library/internal/service/track"
	"lyrics-library/internal/transport/dto"
)

type TrackProvider interface {
	Track(ctx context.Context, artist, title string) (*model.Track, error)
}

type ArtistTracksProvider interface {
	ArtistTracks(ctx context.Context, artist string) ([]*model.Track, error)
}

// @Summary Get song lyrics or artist tracks
// @Description Returns lyrics for specific song or list of all songs by artist if title not provided
// @Tags lyrics
// @Param artist query string true "Artist name" example("Juice WRLD")
// @Param title query string false "Song title (optional)" example("Legends")
// @Success 200 {array} model.Track "Returns artist tracks or track if title provided"
// @Failure 400 {object} dto.ErrorResponse "Artist is required"
// @Failure 404 {object} dto.ErrorResponse "Artist/track not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /lyrics [get]
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

			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "artist is required"})
			return
		}

		if title == "" {
			tracks, err := artistTracksProvider.ArtistTracks(ctx, artist)
			if err != nil {
				if errors.Is(err, trackService.ErrArtistTracksNotFound) {
					c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "artist tracks not found"})
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
				c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "track not found"})
				return
			}

			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
			return
		}

		c.JSON(http.StatusOK, track)
	}
}
