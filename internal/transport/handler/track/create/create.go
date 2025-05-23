package create

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"lyrics-library/internal/lib/logger/sl"
	trackService "lyrics-library/internal/service/track"
	"lyrics-library/internal/transport/dto"
)

type TrackSaver interface {
	Save(ctx context.Context, artist, title string) (*dto.TrackResponse, error)
}

// @Summary Save a new track with translation
// @Description Save lyrics and translation for a given artist and song title
// @Tags track
// @Accept json
// @Produce json
// @Param input body dto.CreateRequest true "Lyrics request data"
// @Success 201 {object} dto.TrackResponse "Successfully saved track"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /track [post]
func New(
	ctx context.Context,
	log *slog.Logger,
	trackSaver TrackSaver,
) gin.HandlerFunc {
	const op = "handler.track.create.New"

	return func(c *gin.Context) {

		log := log.With(slog.String("op", op))

		var req dto.CreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			if errors.Is(err, io.EOF) {
				log.Error("request body is empty")

				c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "request body is empty"})
				return
			}
			log.Error("failed to decode request body", sl.Err(err))

			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request"})
			return
		}

		log.Debug("request body decoded", slog.Any("request", req))

		track, err := trackSaver.Save(ctx, req.Artist, req.Title)
		if err != nil {
			log.Error("failed to create track", sl.Err(err))

			switch {
			case errors.Is(err, trackService.ErrLyricsNotFound):
				c.JSON(http.StatusBadRequest,
					dto.ErrorResponse{Error: "track not found"})
			case errors.Is(err, trackService.ErrFailedTranslateLyrics):
				c.JSON(http.StatusBadRequest,
					dto.ErrorResponse{Error: "failed translate lyrics"})
			default:
				c.JSON(http.StatusInternalServerError,
					dto.ErrorResponse{Error: "internal server error"})
			}
			return
		}

		c.JSON(http.StatusCreated, track)
	}
}
