package delete

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"lyrics-library/internal/service/track"
)

type MockTrackDeleter struct {
	mock.Mock
}

func (m *MockTrackDeleter) Delete(ctx context.Context, uuid string) error {
	args := m.Called(ctx, uuid)
	return args.Error(0)
}

func TestDeleteHandler(t *testing.T) {
	tests := []struct {
		name           string
		uuidParam      string
		mockSetup      func(*MockTrackDeleter)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "successful deletion",
			uuidParam: "123e4567-e89b-12d3-a456-426614174000",
			mockSetup: func(m *MockTrackDeleter) {
				m.On("Delete", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").
					Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
		},
		{
			name:           "missing uuid parameter",
			uuidParam:      "",
			mockSetup:      func(m *MockTrackDeleter) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"uuid is required"}`,
		},
		{
			name:      "invalid uuid format",
			uuidParam: "invalid-uuid",
			mockSetup: func(m *MockTrackDeleter) {
				m.On("Delete", mock.Anything, "invalid-uuid").
					Return(track.ErrInvalidUUID)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid uuid"}`,
		},
		{
			name:      "internal server error",
			uuidParam: "123e4567-e89b-12d3-a456-426614174000",
			mockSetup: func(m *MockTrackDeleter) {
				m.On("Delete", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").
					Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockDeleter := new(MockTrackDeleter)
			tt.mockSetup(mockDeleter)

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			handler := New(context.Background(), log, mockDeleter)

			if tt.uuidParam == "" {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)

				handler(c)

				assert.Equal(t, tt.expectedStatus, w.Code)
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
				return
			}

			router := gin.New()
			router.DELETE("/lyrics/:uuid", handler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodDelete, "/track/"+tt.uuidParam, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			mockDeleter.AssertExpectations(t)
		})
	}
}
