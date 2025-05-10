package login

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"lyrics-library/internal/lib/logger/sl"
	authService "lyrics-library/internal/service/auth"
	"lyrics-library/internal/transport/dto"
)

type UserLogin interface {
	Login(ctx context.Context, credentials *dto.CredentialsRequest) (string, error)
}

// @Summary Login a user
// @Description Login a user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.CredentialsRequest true "Data to login"
// @Success 200 {object} dto.LoginResponse "User logged in successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request data"
// @Failure 401 {object} dto.ErrorResponse "Invalid credentials"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func New(
	ctx context.Context,
	log *slog.Logger,
	userLogin UserLogin,
) gin.HandlerFunc {
	const op = "handler.auth.login.New"

	return func(c *gin.Context) {
		log = log.With(slog.String("op", op))

		var req dto.CredentialsRequest
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

		if err := validator.New().Struct(&req); err != nil {
			log.Error("failed to validate request", sl.Err(err))

			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid email"})
			return
		}

		log.Debug("request body decoded", slog.Any("request", req))

		token, err := userLogin.Login(ctx, &req)
		if err != nil {
			if errors.Is(err, authService.ErrInvalidCredentials) {
				log.Warn("invalid credentials", sl.Err(err))

				c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid credentials"})
				return
			}

			log.Error("failed to login", sl.Err(err))

			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
			return
		}

		c.JSON(http.StatusOK, dto.LoginResponse{Token: token})
	}
}
