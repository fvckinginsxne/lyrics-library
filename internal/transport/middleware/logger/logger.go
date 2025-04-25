package logger

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func New(log *slog.Logger) gin.HandlerFunc {
	log = log.With(
		slog.String("component", "middleware/logger"),
	)

	log.Info("logger middleware enabled")

	return func(c *gin.Context) {
		start := time.Now()

		logEntry := log.With(
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("query", c.Request.URL.RawQuery),
			slog.String("remote_addr", c.Request.RemoteAddr),
			slog.String("user_agent", c.Request.UserAgent()),
			slog.String("request_id", c.GetString("request_id")),
		)

		c.Next()

		latency := time.Since(start)

		logEntry.Info("request completed",
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", latency),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}
