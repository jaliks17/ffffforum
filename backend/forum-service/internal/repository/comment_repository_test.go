package repository

import (
	"context"
	"forum-service/internal/entity"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func setupCommentTest(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	return sqlx.NewDb(db, "sqlmock"), mock
}

func TestCommentRepository_CreateComment(t *testing.T) {
	db, mock := setupCommentTest(t)
	defer db.Close()

	repo := NewCommentRepository(db)
	comment := &entity.Comment{PostID: 1, AuthorID: 1, Content: "Test"}

	mock.ExpectQuery("INSERT INTO comments").
		WithArgs(comment.Content, comment.AuthorID, comment.PostID, comment.AuthorName).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := repo.CreateComment(context.Background(), comment)
	assert.NoError(t, err)
}

func TestCommentRepository_GetCommentByID(t *testing.T) {
	db, mock := setupCommentTest(t)
	defer db.Close()

	repo := NewCommentRepository(db)
	expectedComment := &entity.Comment{ID: 1, Content: "Test", AuthorID: 1, PostID: 1}

	mock.ExpectQuery("SELECT").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "content", "author_id", "post_id", "author_name"}).
			AddRow(expectedComment.ID, expectedComment.Content, expectedComment.AuthorID, expectedComment.PostID, expectedComment.AuthorName))

	comment, err := repo.GetCommentByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, expectedComment, comment)
}

func TestCommentRepository_DeleteComment(t *testing.T) {
	db, mock := setupCommentTest(t)
	defer db.Close()

	repo := NewCommentRepository(db)

	mock.ExpectExec("DELETE FROM comments").
		WithArgs(1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteComment(context.Background(), 1, 1)
	assert.NoError(t, err)
}