package dto

type CreateRequest struct {
	Artist string `json:"artist" binding:"required" example:"Juice WRLD"`
	Title  string `json:"title" binding:"required" example:"Lucid Dreams"`
}
