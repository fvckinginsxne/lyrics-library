package healthchecker

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"lyrics-library/internal/lib/logger/sl"
)

const (
	pdPingTimeout = time.Second * 2
)

type StorageHealthChecker interface {
	Ping(ctx context.Context) error
}

func New(
	log *slog.Logger,
	pgClient StorageHealthChecker,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), pdPingTimeout)
		defer cancel()

		if err := pgClient.Ping(ctx); err != nil {
			log.Error("postgres health check failed", sl.Err(err))

			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"error": "internal server error"},
			)
			return
		}

		c.Next()
	}
}
