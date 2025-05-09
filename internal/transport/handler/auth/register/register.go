package register

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

type UserRegistrar interface {
	Register(ctx context.Context, credentials *dto.RegisterRequest) error
}

// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Param request body dto.RegisterRequest true "Registration data"
// @Success 201 "User created successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request data"
// @Failure 409 {object} dto.ErrorResponse "User already exists"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/register [post]
func New(
	ctx context.Context,
	log *slog.Logger,
	userRegistrar UserRegistrar,
) gin.HandlerFunc {
	const op = "handler.auth.register.New"

	return func(c *gin.Context) {
		log := log.With(slog.String("op", op))

		var req dto.RegisterRequest
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

		if err := userRegistrar.Register(ctx, &req); err != nil {
			if errors.Is(err, authService.ErrUserAlreadyExists) {
				log.Warn("user already exists")

				c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "user already exists"})
				return
			}

			log.Error("failed to register user", sl.Err(err))

			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
			return
		}

		c.Status(http.StatusCreated)
	}
}
