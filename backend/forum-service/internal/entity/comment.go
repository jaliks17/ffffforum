// internal/entity/comment.go
package entity

import "time"

type Comment struct {
	ID         int64     `json:"id" db:"id" example:"1"`
	AuthorID   int64     `json:"author_id" db:"author_id" example:"1"`
	PostID     int64     `json:"post_id" db:"post_id" example:"1"`
	ParentID   *int64    `json:"parent_id" db:"parent_id" example:"1"`
	Content    string    `json:"content" db:"content" example:"текст комментария"`
	CreatedAt  time.Time `db:"created_at"`
	AuthorName string    `json:"author_name" db:"author_name"` // Исправлено db:"-"
}
