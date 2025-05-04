package dto

import "lyrics-library/internal/domain/model"

type ErrorResponse struct {
	Error string `json:"error"`
}

type TrackResponse struct {
	Artist      string   `json:"artist" example:"Lucid Dreams"`
	Title       string   `json:"title" example:"Juice WRLD"`
	Lyrics      []string `json:"track" example:"I still see your shadows in my room..."`
	Translation []string `json:"translation" example:"Я все еще вижу твои тени в моей комнате..."`
}

func ToTrackResponse(t *model.Track) *TrackResponse {
	return &TrackResponse{
		Artist:      t.Artist,
		Title:       t.Title,
		Lyrics:      t.Lyrics,
		Translation: t.Translation,
	}
}

func TracksToTrackResponses(tracks []*model.Track) []*TrackResponse {
	responses := make([]*TrackResponse, len(tracks))

	for i, track := range tracks {
		responses[i] = ToTrackResponse(track)
	}

	return responses
}
