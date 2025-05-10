package login

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

type mockUserLogin struct {
	mock.Mock
}

func (m *mockUserLogin) Login(ctx context.Context, credentials *dto.CredentialsRequest) (string, error) {
	args := m.Called(ctx, credentials)
	return args.String(0), args.Error(1)
}

func TestLoginHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(*mockUserLogin)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful login",
			requestBody: `{"email": "test@example.com", "password": "validpassword123"}`,
			mockSetup: func(m *mockUserLogin) {
				m.On("Login", mock.Anything, &dto.CredentialsRequest{
					Email:    "test@example.com",
					Password: "validpassword123",
				}).Return("jwt.token.here", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"jwt.token.here"`,
		},
		{
			name:           "empty request body",
			requestBody:    "",
			mockSetup:      func(m *mockUserLogin) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"request body is empty"}`,
		},
		{
			name:           "invalid JSON",
			requestBody:    `{"email": "test@example.com", "password": }`,
			mockSetup:      func(m *mockUserLogin) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name:           "missing email",
			requestBody:    `{"password": "validpassword123"}`,
			mockSetup:      func(m *mockUserLogin) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name:           "missing password",
			requestBody:    `{"email": "test@example.com"}`,
			mockSetup:      func(m *mockUserLogin) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name:           "invalid email format",
			requestBody:    `{"email": "notanemail", "password": "validpassword123"}`,
			mockSetup:      func(m *mockUserLogin) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid email"}`,
		},
		{
			name:        "invalid credentials",
			requestBody: `{"email": "wrong@example.com", "password": "wrongpassword"}`,
			mockSetup: func(m *mockUserLogin) {
				m.On("Login", mock.Anything, &dto.CredentialsRequest{
					Email:    "wrong@example.com",
					Password: "wrongpassword",
				}).Return("", authService.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid credentials"}`,
		},
		{
			name:        "internal server error",
			requestBody: `{"email": "test@example.com", "password": "validpassword123"}`,
			mockSetup: func(m *mockUserLogin) {
				m.On("Login", mock.Anything, &dto.CredentialsRequest{
					Email:    "test@example.com",
					Password: "validpassword123",
				}).Return("", errors.New("some internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogin := new(mockUserLogin)
			tt.mockSetup(mockLogin)

			handler := New(
				context.Background(),
				slog.Default(),
				mockLogin,
			)

			req, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			router := gin.New()
			router.POST("/login", handler)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}

			mockLogin.AssertExpectations(t)
		})
	}
}
