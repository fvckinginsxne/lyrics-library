package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"lyrics-library/internal/client/grpc/auth"
	"lyrics-library/internal/transport/dto"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserRegistrar interface {
	Register(ctx context.Context, email, password string) error
}

type Service struct {
	log           *slog.Logger
	userRegistrar UserRegistrar
}

func New(
	log *slog.Logger,
	userRegistrar UserRegistrar,
) *Service {
	return &Service{
		log:           log,
		userRegistrar: userRegistrar,
	}
}

func (s *Service) Register(ctx context.Context, credentials *dto.RegisterRequest) error {
	const op = "service.auth.Register"

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", credentials.Email),
		slog.String("password", credentials.Password),
	)

	log.Info("registering new user")

	err := s.userRegistrar.Register(ctx, credentials.Email, credentials.Password)
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			return fmt.Errorf("%s: %w", op, ErrUserAlreadyExists)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered successfully")

	return nil
}
