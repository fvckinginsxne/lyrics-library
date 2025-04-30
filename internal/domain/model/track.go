package model

type Track struct {
	Artist      string   `json:"artist" example:"Lucid Dreams"`
	Title       string   `json:"title" example:"Juice WRLD"`
	Lyrics      []string `json:"track" example:"I still see your shadows in my room..."`
	Translation []string `json:"translation" example:"Я все еще вижу твои тени в моей комнате..."`
}
