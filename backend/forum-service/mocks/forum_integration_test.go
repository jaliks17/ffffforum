package mocks

import (
	"context"
	"database/sql"
	"fmt"
	"forum-service/internal/entity"
	"forum-service/internal/repository"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

type dbWrapper struct {
	*sqlx.DB
}

func (d *dbWrapper) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	row := d.DB.QueryRowContext(ctx, sql, args...)
	return &pgxRow{row}
}

func (d *dbWrapper) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	result, err := d.DB.ExecContext(ctx, sql, args...)
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	rowsAffected, _ := result.RowsAffected()
	return pgconn.NewCommandTag(fmt.Sprintf("UPDATE %d", rowsAffected)), nil
}

type pgxRow struct {
	*sql.Row
}

func (r *pgxRow) Scan(dest ...interface{}) error {
	return r.Row.Scan(dest...)
}

var testDB *dbWrapper

func TestMain(m *testing.M) {
	// Connect directly using sqlx
	db, err := sqlx.Connect("pgx", "host=localhost port=5432 user=postgres password=postgres dbname=forum_test sslmode=disable")
	if err != nil {
		log.Fatalf("failed to connect to test db: %v", err)
	}
	
	testDB = &dbWrapper{db}
	
	code := m.Run()
	testDB.Close()
	os.Exit(code)
}

func TestPostAndCommentIntegration(t *testing.T) {
	ctx := context.Background()
	postRepo := repository.NewPostRepository(testDB.DB)
	commentRepo := repository.NewCommentRepository(testDB.DB)

	// Create post
	post := &entity.Post{
		Title:    "Integration Post",
		Content:  "Integration Content",
		AuthorID: 1,
	}
	id, err := postRepo.CreatePost(ctx, post)
	assert.NoError(t, err)
	assert.NotZero(t, id)
	post.ID = id

	// Get post
	got, err := postRepo.GetPostByID(ctx, post.ID)
	assert.NoError(t, err)
	assert.Equal(t, post.Title, got.Title)

	// Add comment
	comment := &entity.Comment{
		PostID:     post.ID,
		AuthorID:   1,
		Content:    "Integration Comment",
		AuthorName: "test_user",
	}
	err = commentRepo.CreateComment(ctx, comment)
	assert.NoError(t, err)
	assert.NotZero(t, comment.ID)

	// Get comment
	gotComment, err := commentRepo.GetCommentByID(ctx, comment.ID)
	assert.NoError(t, err)
	assert.Equal(t, comment.Content, gotComment.Content)

	// Delete comment
	err = commentRepo.DeleteComment(ctx, comment.ID, 1)
	assert.NoError(t, err)

	// Delete post
	err = postRepo.DeletePost(ctx, post.ID, 1, "user")
	assert.NoError(t, err)
}

func TestForumIntegration(t *testing.T) {
	t.Skip("Integration tests not implemented yet")
}