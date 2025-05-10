package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	authClient "lyrics-library/internal/client/grpc/auth"
	"lyrics-library/internal/transport/dto"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Auth interface {
	Register(ctx context.Context, email, password string) error
	Login(ctx context.Context, email, password string) (string, error)
}

type Service struct {
	log  *slog.Logger
	auth Auth
}

func New(
	log *slog.Logger,
	auth Auth,
) *Service {
	return &Service{
		log:  log,
		auth: auth,
	}
}

func (s *Service) Register(ctx context.Context, credentials *dto.CredentialsRequest) error {
	const op = "service.auth.Register"

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", credentials.Email),
		slog.String("password", credentials.Password),
	)

	log.Info("registering new user")

	err := s.auth.Register(ctx, credentials.Email, credentials.Password)
	if err != nil {
		if errors.Is(err, authClient.ErrUserAlreadyExists) {
			return fmt.Errorf("%s: %w", op, ErrUserAlreadyExists)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered successfully")

	return nil
}

func (s *Service) Login(ctx context.Context, credentials *dto.CredentialsRequest) (string, error) {
	const op = "service.auth.Login"

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", credentials.Email),
		slog.String("password", credentials.Password),
	)

	log.Info("attempting to login")

	token, err := s.auth.Login(ctx, credentials.Email, credentials.Password)
	if err != nil {
		if errors.Is(err, authClient.ErrInvalidCredentials) {
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully logged in")

	return token, nil
}
