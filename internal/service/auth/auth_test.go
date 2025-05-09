package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"lyrics-library/internal/client/grpc/auth"
	"lyrics-library/internal/transport/dto"
)

type mockUserRegistrar struct {
	mock.Mock
}

func (m *mockUserRegistrar) Register(ctx context.Context, email, password string) error {
	args := m.Called(ctx, email, password)
	return args.Error(0)
}

func TestService_Register(t *testing.T) {
	tests := []struct {
		name        string
		credentials *dto.RegisterRequest
		mockSetup   func(*mockUserRegistrar)
		expectedErr error
	}{
		{
			name: "successful registration",
			credentials: &dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "validpassword123",
			},
			mockSetup: func(m *mockUserRegistrar) {
				m.On("Register", mock.Anything, "test@example.com", "validpassword123").
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "user already exists",
			credentials: &dto.RegisterRequest{
				Email:    "existing@example.com",
				Password: "validpassword123",
			},
			mockSetup: func(m *mockUserRegistrar) {
				m.On("Register", mock.Anything, "existing@example.com", "validpassword123").
					Return(auth.ErrUserAlreadyExists)
			},
			expectedErr: fmt.Errorf("service.auth.Register: %w", ErrUserAlreadyExists),
		},
		{
			name: "internal registrar error",
			credentials: &dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "validpassword123",
			},
			mockSetup: func(m *mockUserRegistrar) {
				m.On("Register", mock.Anything, "test@example.com", "validpassword123").
					Return(errors.New("some internal error"))
			},
			expectedErr: fmt.Errorf("service.auth.Register: %w", errors.New("some internal error")),
		},
		{
			name: "empty email",
			credentials: &dto.RegisterRequest{
				Email:    "",
				Password: "validpassword123",
			},
			mockSetup: func(m *mockUserRegistrar) {
				m.On("Register", mock.Anything, "", "validpassword123").
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "empty password",
			credentials: &dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "",
			},
			mockSetup: func(m *mockUserRegistrar) {
				m.On("Register", mock.Anything, "test@example.com", "").
					Return(nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRegistrar := new(mockUserRegistrar)
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
