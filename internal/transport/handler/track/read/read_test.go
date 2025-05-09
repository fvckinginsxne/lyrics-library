package read

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

	trackService "lyrics-library/internal/service/track"
	"lyrics-library/internal/transport/dto"
)

type MockTrackProvider struct {
	mock.Mock
}

func (m *MockTrackProvider) Track(ctx context.Context, artist, title string) (*dto.TrackResponse, error) {
	args := m.Called(ctx, artist, title)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TrackResponse), args.Error(1)
}

type MockArtistTracksProvider struct {
	mock.Mock
}

func (m *MockArtistTracksProvider) ArtistTracks(ctx context.Context, artist string) ([]*dto.TrackResponse, error) {
	args := m.Called(ctx, artist)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.TrackResponse), args.Error(1)
}

func TestGetHandler(t *testing.T) {
	tests := []struct {
		name               string
		queryParams        map[string]string
		mockTrackProvider  func(*MockTrackProvider)
		mockTracksProvider func(*MockArtistTracksProvider)
		expectedStatus     int
		expectedBody       string
	}{
		{
			name:        "successful track request",
			queryParams: map[string]string{"artist": "Juice WRLD", "title": "Lucid Dreams"},
			mockTrackProvider: func(m *MockTrackProvider) {
				m.On("Track", mock.Anything, "Juice WRLD", "Lucid Dreams").
					Return(&dto.TrackResponse{
						Artist:      "Juice WRLD",
						Title:       "Lucid Dreams",
						Lyrics:      []string{"I still see your shadows in my room..."},
						Translation: []string{"Я все еще вижу твою тени с моей комнате..."},
					}, nil)
			},
			mockTracksProvider: func(m *MockArtistTracksProvider) {},
			expectedStatus:     http.StatusOK,
			expectedBody:       `{"artist":"Juice WRLD","title":"Lucid Dreams","lyrics":["I still see your shadows in my room..."],"translation":["Я все еще вижу твою тени с моей комнате..."]}`,
		},
		{
			name:        "track not found",
			queryParams: map[string]string{"artist": "Unknown", "title": "Nonexistent"},
			mockTrackProvider: func(m *MockTrackProvider) {
				m.On("Track", mock.Anything, "Unknown", "Nonexistent").
					Return(nil, trackService.ErrTrackNotFound)
			},
			mockTracksProvider: func(m *MockArtistTracksProvider) {},
			expectedStatus:     http.StatusBadRequest,
			expectedBody:       `{"error":"track not found"}`,
		},
		{
			name:              "successful artist tracks request",
			queryParams:       map[string]string{"artist": "Juice WRLD"},
			mockTrackProvider: func(m *MockTrackProvider) {},
			mockTracksProvider: func(m *MockArtistTracksProvider) {
				m.On("ArtistTracks", mock.Anything, "Juice WRLD").
					Return([]*dto.TrackResponse{
						{
							Artist:      "Juice WRLD",
							Title:       "Lucid Dreams",
							Lyrics:      []string{"..."},
							Translation: []string{"..."},
						},
						{
							Artist:      "Juice WRLD",
							Title:       "All Girls Are The Same",
							Lyrics:      []string{"..."},
							Translation: []string{"..."},
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"artist":"Juice WRLD","title":"Lucid Dreams","lyrics":["..."],"translation":["..."]},{"artist":"Juice WRLD","title":"All Girls Are The Same","lyrics":["..."],"translation":["..."]}]`,
		},
		{
			name:              "artist tracks not found",
			queryParams:       map[string]string{"artist": "Unknown Artist"},
			mockTrackProvider: func(m *MockTrackProvider) {},
			mockTracksProvider: func(m *MockArtistTracksProvider) {
				m.On("ArtistTracks", mock.Anything, "Unknown Artist").
					Return(nil, trackService.ErrArtistTracksNotFound)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"artist tracks not found"}`,
		},
		{
			name:        "internal server error on track request",
			queryParams: map[string]string{"artist": "Juice WRLD", "title": "Lucid Dreams"},
			mockTrackProvider: func(m *MockTrackProvider) {
				m.On("Track", mock.Anything, "Juice WRLD", "Lucid Dreams").
					Return(nil, errors.New("database error"))
			},
			mockTracksProvider: func(m *MockArtistTracksProvider) {},
			expectedStatus:     http.StatusInternalServerError,
			expectedBody:       `{"error":"internal server error"}`,
		},
		{
			name:              "internal server error on artist tracks request",
			queryParams:       map[string]string{"artist": "Juice WRLD"},
			mockTrackProvider: func(m *MockTrackProvider) {},
			mockTracksProvider: func(m *MockArtistTracksProvider) {
				m.On("ArtistTracks", mock.Anything, "Juice WRLD").
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
		{
			name:               "missing artist parameter",
			queryParams:        map[string]string{"title": "Lucid Dreams"},
			mockTrackProvider:  func(m *MockTrackProvider) {},
			mockTracksProvider: func(m *MockArtistTracksProvider) {},
			expectedStatus:     http.StatusBadRequest,
			expectedBody:       `{"error":"artist is required"}`,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockTrackProvider := new(MockTrackProvider)
			mockTracksProvider := new(MockArtistTracksProvider)

			tt.mockTrackProvider(mockTrackProvider)
			tt.mockTracksProvider(mockTracksProvider)

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			handler := New(context.Background(), log, mockTrackProvider, mockTracksProvider)

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()
			ctx.Request = req

			handler(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())

			mockTrackProvider.AssertExpectations(t)
			mockTracksProvider.AssertExpectations(t)
		})
	}
}
