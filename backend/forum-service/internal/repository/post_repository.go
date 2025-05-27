package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jaliks17/ffffforum/backend/forum-service/internal/entity"

	"github.com/jmoiron/sqlx"
)

var (
	ErrPostNotFound     = errors.New("post not found")
	ErrPermissionDenied = errors.New("permission denied")
)

type PostRepository interface {
	CreatePost(ctx context.Context, post *entity.Post) (int64, error)
	GetPosts(ctx context.Context) ([]*entity.Post, error)
	GetPostByID(ctx context.Context, id int64) (*entity.Post, error)
	DeletePost(ctx context.Context, id, authorID int64, role string) error
	UpdatePost(ctx context.Context, id, authorID int64, role, title, content string) (*entity.Post, error)
}

type postRepository struct {
	db *sqlx.DB
}

func NewPostRepository(db *sqlx.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) CreatePost(ctx context.Context, post *entity.Post) (int64, error) {
	query := `
		INSERT INTO posts (title, content, author_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	var id int64
	err := r.db.QueryRowContext(ctx, query,
		post.Title,
		post.Content,
		post.AuthorID,
		post.CreatedAt,
	).Scan(&id)

	return id, err
}

func (r *postRepository) GetPosts(ctx context.Context) ([]*entity.Post, error) {
	query := `
		SELECT 
			id,
			title,
			content,
			author_id,
			created_at
		FROM posts
		ORDER BY created_at DESC`

	var posts []*entity.Post
	err := r.db.SelectContext(ctx, &posts, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*entity.Post{}, nil
		}
		return nil, err
	}
	return posts, nil
}

func (r *postRepository) GetPostByID(ctx context.Context, id int64) (*entity.Post, error) {
	query := `
		SELECT 
			id,
			title,
			content,
			author_id,
			created_at
		FROM posts
		WHERE id = $1`

	var post entity.Post
	err := r.db.GetContext(ctx, &post, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) DeletePost(ctx context.Context, id, authorID int64, role string) error {
	query := `
		DELETE FROM posts 
		WHERE id = $1 
		AND (author_id = $2 OR $3 = 'admin')`

	result, err := r.db.ExecContext(ctx, query, id, authorID, role)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrPostNotFound
	}

	return nil
}

func (r *postRepository) UpdatePost(ctx context.Context, id, authorID int64, role, title, content string) (*entity.Post, error) {
	query := `
		UPDATE posts
		SET title = $1, content = $2
		WHERE id = $3 AND (author_id = $4 OR $5 = 'admin')
		RETURNING id, title, content, author_id, created_at`

	var post entity.Post
	err := r.db.QueryRowContext(ctx, query,
		title,
		content,
		id,
		authorID,
		role,
	).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.AuthorID,
		&post.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	return &post, nil
}