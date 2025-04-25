package delete

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"lyrics-library/internal/service/track"
)

type TrackDeleter interface {
	Delete(ctx context.Context, uuid string) error
}

func New(
	ctx context.Context,
	log *slog.Logger,
	trackDeleter TrackDeleter,
) gin.HandlerFunc {
	const op = "handlers.song.delete.New"

	return func(c *gin.Context) {
		log = log.With("op", op)

		log.Info("deleting track")

		uuid := c.Param("uuid")
		if uuid == "" {
			log.Error("uuid is required")

			c.JSON(http.StatusBadRequest, gin.H{"error": "uuid is required"})
			return
		}

		if err := trackDeleter.Delete(ctx, uuid); err != nil {
			if errors.Is(err, track.ErrInvalidUUID) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.Status(http.StatusOK)
	}
}
