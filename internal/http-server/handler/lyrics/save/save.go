package save

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"lyrics-library/internal/domain/models"
	"lyrics-library/internal/lib/logger/sl"
	trackService "lyrics-library/internal/service/track"
)

type Request struct {
	Artist string `json:"artist" validate:"required"`
	Title  string `json:"title" validate:"required"`
}

type TrackSaver interface {
	Save(ctx context.Context, artist, title string) (*models.Track, error)
}

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

		log.Info("saving lyrics")

		var req Request
		if err := c.ShouldBindJSON(&req); err != nil {
			if errors.Is(err, io.EOF) {
				log.Error("request body is empty")

				c.JSON(http.StatusBadRequest, gin.H{"error": "request body is empty"})
				return
			}
			log.Error("failed to decode request body", sl.Err(err))

			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		log.Debug("request body decoded", slog.Any("request", req))

		track, err := trackSaver.Save(ctx, req.Artist, req.Title)
		if err != nil {
			log.Error("failed to save lyrics", sl.Err(err))

			switch {
			case errors.Is(err, trackService.ErrLyricsNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": "lyrics not found"})
			case errors.Is(err, trackService.ErrFailedTranslateLyrics):
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed translate lyrics"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
			return
		}

		c.JSON(http.StatusCreated, track)
	}
}
