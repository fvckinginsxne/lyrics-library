package save

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"lyrics-library/internal/domain/model"
	"lyrics-library/internal/lib/logger/sl"
	trackService "lyrics-library/internal/service/track"
	"lyrics-library/internal/transport/dto"
)

type Request struct {
	Artist string `json:"artist" binding:"required" example:"Juice WRLD"`
	Title  string `json:"title" binding:"required" example:"Legends"`
}

type TrackSaver interface {
	Save(ctx context.Context, artist, title string) (*model.Track, error)
}

// @Summary Save new track with translation
// @Description Saves track and translation for a given artist and song title
// @Tags track
// @Accept json
// @Produce json
// @Param input body Request true "Lyrics request data"
// @Success 201 {object} dto.TrackResponse "Successfully saved track"
// @Failure 400 {object} dto.ErrorResponse "Invalid request data"
// @Failure 404 {object} dto.ErrorResponse "Lyrics not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /track [post]
func New(
	ctx context.Context,
	log *slog.Logger,
	trackSaver TrackSaver,
) gin.HandlerFunc {
	const op = "handlers.song.save.New"

	return func(c *gin.Context) {

		log := log.With(
			slog.String("op", op),
		)

		var req Request
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
			log.Error("failed to save track", sl.Err(err))

			switch {
			case errors.Is(err, trackService.ErrLyricsNotFound):
				c.JSON(http.StatusNotFound,
					dto.ErrorResponse{Error: "track not found"})
			case errors.Is(err, trackService.ErrFailedTranslateLyrics):
				c.JSON(http.StatusInternalServerError,
					dto.ErrorResponse{Error: "failed translate track"})
			default:
				c.JSON(http.StatusInternalServerError,
					dto.ErrorResponse{Error: "internal server error"})
			}
			return
		}

		c.JSON(http.StatusCreated, dto.ToTrackResponse(track))
	}
}
