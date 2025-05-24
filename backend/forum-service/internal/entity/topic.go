package entity

import (
	"time"
)

type Topic struct {
	ID         int64     `db:"id"`
	CategoryID int64     `db:"category_id"`
	Title      string    `db:"title"`
	UserID     int64     `db:"user_id"`
	CreatedAt  time.Time `db:"created_at"`
}








