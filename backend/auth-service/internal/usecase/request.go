package usecase

import "time"

type RegisterRequest struct {
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}

type LoginRequest struct {
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}

type ValidateTokenRequest struct {
	Token string
}

type GetUserRequest struct {
	UserID int64
}
type Config struct {
	TokenSecret     string
	TokenExpiration time.Duration
}