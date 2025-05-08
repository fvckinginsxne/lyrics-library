package delete

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	trackService "lyrics-library/internal/service/track"
	"lyrics-library/internal/transport/dto"
)

type TrackDeleter interface {
	Delete(ctx context.Context, uuid string) error
}

// @Summary Delete song track
// @Description Delete song track by uuid
// @Tags track
// @Param uuid path string true "Track UUID" example(e434dc13-ada5-4bde-b695-d97014dadebc)
// @Success 204 "Track deleted successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /track/{uuid} [delete]
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

			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "uuid is required"})
			return
		}

		if err := trackDeleter.Delete(ctx, uuid); err != nil {
			if errors.Is(err, trackService.ErrInvalidUUID) {

				c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "track not found"})
				return
			}

			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
			return
		}

		c.Status(http.StatusNoContent)
	}
}
