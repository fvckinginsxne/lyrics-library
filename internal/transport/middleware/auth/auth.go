package auth

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"lyrics-library/internal/client/grpc/auth"
	"lyrics-library/internal/transport/dto"
)

func New(
	log *slog.Logger,
	authClient *auth.Client,
) gin.HandlerFunc {
	log = log.With(
		slog.String("component", "middleware/auth"),
	)

	log.Info("auth middleware enabled")

	return func(c *gin.Context) {
		token, err := c.Cookie("jwt")
		if err != nil {
			log.Warn("token cookie not found")

			return
		}

		log.Info("token cookie found")

		uid, err := authClient.ParseToken(c.Request.Context(), token)
		if err != nil {
			log.Error("provided invalid token")

			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid token"})
			return
		}

		log.Info("token is valid")

		c.Set("uid", uid)
		c.Next()
	}
}
