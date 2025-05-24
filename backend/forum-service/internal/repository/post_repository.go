package repository

import (
	"context"
	"errors"
	"forum-service/internal/entity"
	"time"

	"github.com/jmoiron/sqlx"
)

// Error constants should be defined in errors.go
// var (
// 	ErrPostNotFound     = errors.New("post not found")
// 	ErrPermissionDenied = errors.New("permission denied")
// )

type PostRepository interface {
	CreatePost(ctx context.Context, post *entity.Post) (int64, error)
	GetPosts(ctx context.Context) ([]*entity.Post, error)
	GetPostByID(ctx context.Context, id int64) (*entity.Post, error)
	DeletePost(ctx context.Context, postID, authorID int64, role string) error
	UpdatePost(ctx context.Context, postID, authorID int64, role, title, content string) (*entity.Post, error)
	GetPostsByUserID(ctx context.Context, userID int64) ([]*entity.Post, error)
}

type PostRepositoryImpl struct {
	db *sqlx.DB
}

func NewPostRepository(db *sqlx.DB) PostRepository {
	return &PostRepositoryImpl{db: db}
}

func (r *PostRepositoryImpl) CreatePost(ctx context.Context, post *entity.Post) (int64, error) {
	query := `INSERT INTO posts (title, content, author_id, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var id int64
	err := r.db.QueryRowContext(ctx, query,
		post.Title, post.Content, post.AuthorID, time.Now(), time.Now(),
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *PostRepositoryImpl) GetPostByID(ctx context.Context, id int64) (*entity.Post, error) {
	query := `SELECT id, title, content, author_id, created_at, updated_at FROM posts WHERE id = $1`
	post := &entity.Post{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.CreatedAt, &post.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (r *PostRepositoryImpl) DeletePost(ctx context.Context, postID, authorID int64, role string) error {
	// This method needs implementation based on your database schema and logic for deletion with user/role check
	return errors.New("DeletePost not implemented") // Placeholder
}

func (r *PostRepositoryImpl) UpdatePost(ctx context.Context, postID, authorID int64, role, title, content string) (*entity.Post, error) {
	// This method needs implementation based on your database schema and logic for updating with user/role check
	return nil, errors.New("UpdatePost not implemented") // Placeholder
}

func (r *PostRepositoryImpl) GetPosts(ctx context.Context) ([]*entity.Post, error) {
	query := `SELECT id, title, content, author_id, created_at, updated_at FROM posts ORDER BY created_at DESC`

	var posts []*entity.Post
	err := r.db.SelectContext(ctx, &posts, query)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *PostRepositoryImpl) GetPostsByUserID(ctx context.Context, userID int64) ([]*entity.Post, error) {
	// This method needs implementation to fetch posts by user ID
	return nil, errors.New("GetPostsByUserID not implemented") // Placeholder
}
