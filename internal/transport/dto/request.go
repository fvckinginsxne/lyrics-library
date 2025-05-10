package dto

type CreateRequest struct {
	Artist string `json:"artist" binding:"required" example:"Juice WRLD"`
	Title  string `json:"title" binding:"required" example:"Lucid Dreams"`
}

type CredentialsRequest struct {
	Email    string `json:"email" binding:"required" validate:"email" example:"test@test.com"`
	Password string `json:"password" binding:"required" example:"matveyisgoat123"`
}
