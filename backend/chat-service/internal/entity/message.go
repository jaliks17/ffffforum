package entity

import "time"

type Message struct {
	ID         int       `json:"id" example:"1" db:"id"`
	UserID     int       `json:"user_id" example:"123" db:"user_id"`
	Username   string    `json:"username" example:"john_doe" db:"username"`
	Message    string    `json:"message" example:"Hello, world!" db:"content"`
	Timestamp  time.Time `json:"timestamp" example:"2023-10-27T10:00:00Z" db:"timestamp"`
}