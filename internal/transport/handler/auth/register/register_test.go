package register

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	authService "lyrics-library/internal/service/auth"
	"lyrics-library/internal/transport/dto"
)

type mockUserRegistrar struct {
	mock.Mock
}

func (m *mockUserRegistrar) Register(ctx context.Context, credentials *dto.RegisterRequest) error {
	args := m.Called(ctx, credentials)
	return args.Error(0)
}

func TestRegisterHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(*mockUserRegistrar)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful registration",
			requestBody: `{"email": "test@example.com", "password": "validpassword123"}`,
			mockSetup: func(m *mockUserRegistrar) {
				m.On("Register", mock.Anything, &dto.RegisterRequest{
					Email:    "test@example.com",
					Password: "validpassword123",
				}).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "",
		},
		{
			name:           "empty request body",
			requestBody:    "",
			mockSetup:      func(m *mockUserRegistrar) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"request body is empty"}`,
		},
		{
			name:           "invalid JSON",
			requestBody:    `{"email": "test@example.com", "password": }`,
			mockSetup:      func(m *mockUserRegistrar) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name:           "missing email",
			requestBody:    `{"password": "validpassword123"}`,
			mockSetup:      func(m *mockUserRegistrar) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name:           "missing password",
			requestBody:    `{"email": "test@example.com"}`,
			mockSetup:      func(m *mockUserRegistrar) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name:           "invalid email format",
			requestBody:    `{"email": "notanemail", "password": "validpassword123"}`,
			mockSetup:      func(m *mockUserRegistrar) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid email"}`,
		},
		{
			name:        "user already exists",
			requestBody: `{"email": "existing@example.com", "password": "validpassword123"}`,
			mockSetup: func(m *mockUserRegistrar) {
				m.On("Register", mock.Anything, &dto.RegisterRequest{
					Email:    "existing@example.com",
					Password: "validpassword123",
				}).Return(authService.ErrUserAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"user already exists"}`,
		},
		{
			name:        "internal server error",
			requestBody: `{"email": "test@example.com", "password": "validpassword123"}`,
			mockSetup: func(m *mockUserRegistrar) {
				m.On("Register", mock.Anything, &dto.RegisterRequest{
					Email:    "test@example.com",
					Password: "validpassword123",
				}).Return(errors.New("some internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRegistrar := new(mockUserRegistrar)
			tt.mockSetup(mockRegistrar)

			handler := New(
				context.Background(),
				slog.Default(),
				mockRegistrar,
			)

			req, _ := http.NewRequest(http.MethodPost, "/register", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			router := gin.New()
			router.POST("/register", handler)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			} else {
				assert.Empty(t, w.Body.String())
			}

			mockRegistrar.AssertExpectations(t)
		})
	}
}
