package entity

type ErrorResponse struct {
	Error string `json:"error" example:"Internal server error"`
}