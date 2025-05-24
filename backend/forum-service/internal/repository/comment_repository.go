package repository

import (
	"context"
	"database/sql"
	"errors"
	"forum-service/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrCommentPermissionDenied = errors.New("permission denied")
)

type DB interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

type CommentRepository interface {
	CreateComment(ctx context.Context, comment *entity.Comment) error
	GetCommentsByPostID(ctx context.Context, postID int64) ([]entity.Comment, error)
	GetCommentByID(ctx context.Context, id int64) (*entity.Comment, error)
	DeleteComment(ctx context.Context, id int64, userID int64) error
}

type CommentRepo struct {
	db *sqlx.DB
}

func NewCommentRepository(db *sqlx.DB) CommentRepository {
	return &CommentRepo{db: db}
}

func (r *CommentRepo) CreateComment(ctx context.Context, comment *entity.Comment) error {
	query := `INSERT INTO comments (content, author_id, post_id, author_name) 
        VALUES ($1, $2, $3, $4) RETURNING id`
	return r.db.QueryRowContext(ctx, query,
		comment.Content,
		comment.AuthorID,
		comment.PostID,
		comment.AuthorName,
	).Scan(&comment.ID)
}

func (r *CommentRepo) GetCommentsByPostID(ctx context.Context, postID int64) ([]entity.Comment, error) {
	query := `
        SELECT 
            id,
            content,
            author_id,
            post_id,
            author_name
        FROM comments 
        WHERE post_id = $1
        ORDER BY id DESC`

	var comments []entity.Comment
	err := r.db.SelectContext(ctx, &comments, query, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []entity.Comment{}, nil
		}
		return nil, err
	}
	return comments, nil
}

func (r *CommentRepo) GetCommentByID(ctx context.Context, id int64) (*entity.Comment, error) {
	query := `
        SELECT 
            id,
            content,
            author_id,
            post_id,
            author_name
        FROM comments 
        WHERE id = $1`

	var comment entity.Comment
	err := r.db.GetContext(ctx, &comment, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepo) DeleteComment(ctx context.Context, id int64, userID int64) error {
	// Проверяем, существует ли комментарий и принадлежит ли он пользователю
	query := `DELETE FROM comments WHERE id = $1 AND author_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// Если затронуто 0 строк, это может означать, что либо комментария с таким ID нет,
	// либо он не принадлежит указанному пользователю.
	// Чтобы различить эти случаи, сначала проверим существование комментария.
	
	if rowsAffected == 0 {
		// Попробуем получить комментарий по ID, чтобы узнать, существует ли он.
		_, err := r.GetCommentByID(ctx, id)
		if err != nil {
			// Если GetCommentByID вернул ошибку (например, ErrCommentNotFound), возвращаем ее.
			return err
		}
		// Если GetCommentByID не вернул ошибку, значит комментарий существует,
		// но не принадлежит текущему пользователю.
		return ErrCommentPermissionDenied
	}

	return nil
}