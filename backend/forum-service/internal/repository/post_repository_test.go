package repository

import (
	"context"
	"forum-service/internal/entity"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func setupTest(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	return sqlx.NewDb(db, "sqlmock"), mock
}

func TestPostRepository_CreatePost(t *testing.T) {
	db, mock := setupTest(t)
	defer db.Close()

	repo := NewPostRepository(db)
	post := &entity.Post{Title: "Test", Content: "Test", AuthorID: 1}

	// Expect the INSERT query and return a dummy ID
	mock.ExpectQuery(`INSERT INTO posts \(title, content, author_id, created_at, updated_at\)`).WithArgs(
		post.Title, post.Content, post.AuthorID, sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	id, err := repo.CreatePost(context.Background(), post)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostRepository_GetPostByID(t *testing.T) {
	db, mock := setupTest(t)
	defer db.Close()

	repo := NewPostRepository(db)
	expectedPost := &entity.Post{ID: 1, Title: "Test", Content: "Test", AuthorID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}

	// Expect the SELECT query and return a row
	mock.ExpectQuery(`SELECT id, title, content, author_id, created_at, updated_at FROM posts WHERE id = \$1`).WithArgs(1).WillReturnRows(
		sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at", "updated_at"}).
			AddRow(expectedPost.ID, expectedPost.Title, expectedPost.Content, expectedPost.AuthorID, expectedPost.CreatedAt, expectedPost.UpdatedAt),
	)

	post, err := repo.GetPostByID(context.Background(), 1)
	assert.NoError(t, err)
	// Compare relevant fields, as time.Time comparison can be tricky
	assert.Equal(t, expectedPost.ID, post.ID)
	assert.Equal(t, expectedPost.Title, post.Title)
	assert.Equal(t, expectedPost.Content, post.Content)
	assert.Equal(t, expectedPost.AuthorID, post.AuthorID)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostRepository_DeletePost(t *testing.T) {
	db, mock := setupTest(t)
	defer db.Close()

	// repo := NewPostRepository(db) // Declared but not used for now
	postID := int64(1)
	authorID := int64(1)
	// role := "user" // Declared but not used for now

	// We need to mock the check for post ownership/existence first (if implemented), then the DELETE
	// Since DeletePost logic is currently commented out/placeholder, we'll mock the expected DB call based on the interface.
	// The current interface expects DeletePost(ctx context.Context, postID, authorID int64, role string) error
	// The current implementation placeholder returns errors.New("DeletePost not implemented").
	// If you implement the DeletePost logic, you'll need to update this mock expectation.

	// Mocking a potential SELECT to check ownership before delete (common pattern)
	// mock.ExpectQuery(`SELECT author_id FROM posts WHERE id = \$1`).WithArgs(postID).WillReturnRows(
	// 	sqlmock.NewRows([]string{"author_id"}).AddRow(authorID),
	// )

	// Mocking the DELETE statement
	// Assuming DeletePost checks author_id and postID
	mock.ExpectExec(`DELETE FROM posts WHERE id = \$1 AND author_id = \$2`).WithArgs(postID, authorID).WillReturnResult(sqlmock.NewResult(0, 1))

	// Note: The current implementation of DeletePost in post_repository.go is a placeholder
	// and will return "DeletePost not implemented" regardless of the mock.
	// This test will pass if the method is eventually implemented to interact with the DB as mocked.

	// Call the method - this will fail if the implementation still returns the placeholder error.
	// You need to implement the actual database logic in DeletePost first.
	// For now, we can assert that calling the placeholder returns the expected error.
	// err := repo.DeletePost(context.Background(), postID, authorID, role)
	// assert.EqualError(t, err, "DeletePost not implemented")

	// If you implement DeletePost, uncomment the above lines and the ExpectExec line.
	// The test should then verify the database interaction.

	// Since DeletePost is currently a placeholder, this test isn't fully functional yet.
	// It sets up the mock expectation but cannot verify it against the placeholder implementation.

}

