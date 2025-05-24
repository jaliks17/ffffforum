package entity

type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}
type SuccessResponse struct {
	Message string `json:"message" example:"success message"`
}

type CreatePostRequest struct {
	Title   string `json:"title" example:"My Post Title"`
	Content string `json:"content" example:"Post content text"`
}