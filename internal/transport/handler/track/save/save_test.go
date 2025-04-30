package save

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"lyrics-library/internal/domain/model"
	trackService "lyrics-library/internal/service/track"
)

type MockTrackSaver struct {
	mock.Mock
}

func (m *MockTrackSaver) Save(ctx context.Context, artist, title string) (*model.Track, error) {
	args := m.Called(ctx, artist, title)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Track), args.Error(1)
}

func TestSaveHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(*MockTrackSaver)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful save",
			requestBody: `{"artist": "Juice WRLD", "title": "Lucid Dreams"}`,
			mockSetup: func(m *MockTrackSaver) {
				m.On("Save", mock.Anything, "Juice WRLD", "Lucid Dreams").
					Return(&model.Track{
						Artist:      "Juice WRLD",
						Title:       "Lucid Dreams",
						Lyrics:      []string{"I still see your shadows in my room..."},
						Translation: []string{"Я все еще вижу твою тени с моей комнате..."},
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"artist":"Juice WRLD","title":"Lucid Dreams","track":["I still see your shadows in my room..."],"translation":["Я все еще вижу твою тени с моей комнате..."]}`,
		},
		{
			name:           "empty request body",
			requestBody:    "",
			mockSetup:      func(m *MockTrackSaver) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"request body is empty"}`,
		},
		{
			name:           "invalid json",
			requestBody:    `{"artist": "Juice WRLD", "title": }`,
			mockSetup:      func(m *MockTrackSaver) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name:        "track not found",
			requestBody: `{"artist": "Juice WRLD", "title": "Lucid Dreams"}`,
			mockSetup: func(m *MockTrackSaver) {
				m.On("Save", mock.Anything, "Juice WRLD", "Lucid Dreams").
					Return(nil, trackService.ErrLyricsNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"track not found"}`,
		},
		{
			name:        "failed to translate track",
			requestBody: `{"artist": "Juice WRLD", "title": "Lucid Dreams"}`,
			mockSetup: func(m *MockTrackSaver) {
				m.On("Save", mock.Anything, "Juice WRLD", "Lucid Dreams").
					Return(nil, trackService.ErrFailedTranslateLyrics)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed translate track"}`,
		},
		{
			name:        "internal server error",
			requestBody: `{"artist": "Juice WRLD", "title": "Lucid Dreams"}`,
			mockSetup: func(m *MockTrackSaver) {
				m.On("Save", mock.Anything, "Juice WRLD", "Lucid Dreams").
					Return(nil, errors.New("some unexpected error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSaver := new(MockTrackSaver)
			tt.mockSetup(mockSaver)

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			handler := New(context.Background(), log, mockSaver)

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			var req *http.Request
			if tt.requestBody != "" {
				req = httptest.NewRequest("POST", "/", strings.NewReader(tt.requestBody))
			} else {
				req = httptest.NewRequest("POST", "/", nil)
			}
			ctx.Request = req

			handler(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())

			mockSaver.AssertExpectations(t)
		})
	}
}
