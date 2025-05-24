package entity

import "time"

type Post struct {
	ID        int64     `json:"id" db:"id" example:"123"`
	Title     string    `json:"title" db:"title" example:"My Post Title"`
	Content   string    `json:"content" db:"content" example:"Post content text"`
	AuthorID  int64     `json:"author_id" db:"author_id" example:"456"`
	CreatedAt time.Time `json:"created_at" db:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" example:"2023-01-01T00:00:00Z"`
}
