package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	authClient "lyrics-library/internal/client/grpc/auth"
	"lyrics-library/internal/transport/dto"
)

type mockAuth struct {
	mock.Mock
}

func (m *mockAuth) Register(ctx context.Context, email, password string) error {
	args := m.Called(ctx, email, password)
	return args.Error(0)
}

func (m *mockAuth) Login(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func TestService_Register(t *testing.T) {
	tests := []struct {
		name        string
		credentials *dto.CredentialsRequest
		mockSetup   func(*mockAuth)
		expectedErr error
	}{
		{
			name: "successful registration",
			credentials: &dto.CredentialsRequest{
				Email:    "test@example.com",
				Password: "validpassword123",
			},
			mockSetup: func(m *mockAuth) {
				m.On("Register", mock.Anything, "test@example.com", "validpassword123").
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "user already exists",
			credentials: &dto.CredentialsRequest{
				Email:    "existing@example.com",
				Password: "validpassword123",
			},
			mockSetup: func(m *mockAuth) {
				m.On("Register", mock.Anything, "existing@example.com", "validpassword123").
					Return(authClient.ErrUserAlreadyExists)
			},
			expectedErr: fmt.Errorf("service.auth.Register: %w", ErrUserAlreadyExists),
		},
		{
			name: "internal registrar error",
			credentials: &dto.CredentialsRequest{
				Email:    "test@example.com",
				Password: "validpassword123",
			},
			mockSetup: func(m *mockAuth) {
				m.On("Register", mock.Anything, "test@example.com", "validpassword123").
					Return(errors.New("some internal error"))
			},
			expectedErr: fmt.Errorf("service.auth.Register: %w", errors.New("some internal error")),
		},
		{
			name: "empty email",
			credentials: &dto.CredentialsRequest{
				Email:    "",
				Password: "validpassword123",
			},
			mockSetup: func(m *mockAuth) {
				m.On("Register", mock.Anything, "", "validpassword123").
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "empty password",
			credentials: &dto.CredentialsRequest{
				Email:    "test@example.com",
				Password: "",
			},
			mockSetup: func(m *mockAuth) {
				m.On("Register", mock.Anything, "test@example.com", "").
					Return(nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRegistrar := new(mockAuth)
			tt.mockSetup(mockRegistrar)

			service := New(slog.Default(), mockRegistrar)

			err := service.Register(context.Background(), tt.credentials)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRegistrar.AssertExpectations(t)
		})
	}
}

func TestService_Login(t *testing.T) {
	tests := []struct {
		name        string
		credentials *dto.CredentialsRequest
		mockSetup   func(*mockAuth)
		expectedErr error
	}{
		{
			name: "successful login",
			credentials: &dto.CredentialsRequest{
				Email:    "test@example.com",
				Password: "validpassword123",
			},
			mockSetup: func(m *mockAuth) {
				m.On("Login", mock.Anything, "test@example.com", "validpassword123").
					Return("jwt.token", nil)
			},
			expectedErr: nil,
		},
		{
			name: "invalid credentials",
			credentials: &dto.CredentialsRequest{
				Email:    "invalid@example.com",
				Password: "invalidpassword123",
			},
			mockSetup: func(m *mockAuth) {
				m.On("Login", mock.Anything, "invalid@example.com", "invalidpassword123").
					Return("", authClient.ErrInvalidCredentials)
			},
			expectedErr: fmt.Errorf("service.auth.Login: %w", ErrInvalidCredentials),
		},
		{
			name: "internal login error",
			credentials: &dto.CredentialsRequest{
				Email:    "test@example.com",
				Password: "validpassword123",
			},
			mockSetup: func(m *mockAuth) {
				m.On("Login", mock.Anything, "test@example.com", "validpassword123").
					Return("", errors.New("some internal error"))
			},
			expectedErr: fmt.Errorf("service.auth.Login: %w", errors.New("some internal error")),
		},
		{
			name: "empty email",
			credentials: &dto.CredentialsRequest{
				Email:    "",
				Password: "validpassword123",
			},
			mockSetup: func(m *mockAuth) {
				m.On("Login", mock.Anything, "", "validpassword123").
					Return("jwt.token", nil)
			},
			expectedErr: nil,
		},
		{
			name: "empty password",
			credentials: &dto.CredentialsRequest{
				Email:    "test@example.com",
				Password: "",
			},
			mockSetup: func(m *mockAuth) {
				m.On("Login", mock.Anything, "test@example.com", "").
					Return("jwt.token", nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogin := new(mockAuth)
			tt.mockSetup(mockLogin)

			service := New(slog.Default(), mockLogin)

			_, err := service.Login(context.Background(), tt.credentials)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockLogin.AssertExpectations(t)
		})
	}
}
