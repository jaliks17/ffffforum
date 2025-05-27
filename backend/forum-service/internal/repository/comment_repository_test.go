package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jaliks17/ffffforum/backend/forum-service/internal/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestCreateComment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCommentRepository(sqlxDB)

	tests := []struct {
		name    string
		comment *entity.Comment
		mock    func()
		wantID  int64
		wantErr bool
	}{
		{
			name: "Success",
			comment: &entity.Comment{
				Content:    "Test comment",
				AuthorID:   1,
				PostID:     1,
				AuthorName: "testuser",
			},
			mock: func() {
				mock.ExpectQuery(`INSERT INTO comments`).
					WithArgs("Test comment", int64(1), int64(1), "testuser").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			wantID: 1,
		},
		{
			name: "Empty Content",
			comment: &entity.Comment{
				Content:    "",
				AuthorID:   1,
				PostID:     1,
				AuthorName: "testuser",
			},
			mock: func() {
				mock.ExpectQuery(`INSERT INTO comments`).
					WithArgs("", int64(1), int64(1), "testuser").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			err := repo.CreateComment(context.Background(), tt.comment)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.comment.ID != tt.wantID {
				t.Errorf("CreateComment() got ID = %v, want %v", tt.comment.ID, tt.wantID)
			}
		})
	}
}

func TestGetCommentsByPostID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCommentRepository(sqlxDB)

	tests := []struct {
		name    string
		postID  int64
		mock    func()
		want    []entity.Comment
		wantErr bool
	}{
		{
			name:   "Success",
			postID: 1,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "content", "author_id", "post_id", "author_name"}).
					AddRow(1, "Comment 1", 1, 1, "user1").
					AddRow(2, "Comment 2", 2, 1, "user2")
				mock.ExpectQuery(`SELECT`).WithArgs(int64(1)).WillReturnRows(rows)
			},
			want: []entity.Comment{
				{
					ID:         1,
					Content:    "Comment 1",
					AuthorID:   1,
					PostID:     1,
					AuthorName: "user1",
				},
				{
					ID:         2,
					Content:    "Comment 2",
					AuthorID:   2,
					PostID:     1,
					AuthorName: "user2",
				},
			},
		},
		{
			name:   "No Comments",
			postID: 2,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "content", "author_id", "post_id", "author_name"})
				mock.ExpectQuery(`SELECT`).WithArgs(int64(2)).WillReturnRows(rows)
			},
			want:    []entity.Comment{},
			wantErr: false,
		},
		{
			name:   "Database Error",
			postID: 3,
			mock: func() {
				mock.ExpectQuery(`SELECT`).WithArgs(int64(3)).WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.GetCommentsByPostID(context.Background(), tt.postID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCommentsByPostID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if len(tt.want) == 0 {
				if got != nil {
					assert.Len(t, got, 0)
				} else {
					assert.Nil(t, got)
				}
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetCommentByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCommentRepository(sqlxDB)

	tests := []struct {
		name     string
		commentID int64
		mock     func()
		want     *entity.Comment
		wantErr  error
	}{
		{
			name: "Success",
			commentID: 1,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "content", "author_id", "post_id", "author_name"}).
					AddRow(1, "Test Comment", 1, 1, "testuser")
				mock.ExpectQuery(`SELECT`).WithArgs(int64(1)).WillReturnRows(rows)
			},
			want: &entity.Comment{
				ID:         1,
				Content:    "Test Comment",
				AuthorID:   1,
				PostID:     1,
				AuthorName: "testuser",
			},
			wantErr: nil,
		},
		{
			name: "Not Found",
			commentID: 2,
			mock: func() {
				mock.ExpectQuery(`SELECT`).WithArgs(int64(2)).WillReturnError(sql.ErrNoRows)
			},
			want: nil,
			wantErr: ErrCommentNotFound,
		},
		{
			name: "Database Error",
			commentID: 3,
			mock: func() {
				mock.ExpectQuery(`SELECT`).WithArgs(int64(3)).WillReturnError(sql.ErrConnDone)
			},
			want: nil,
			wantErr: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.GetCommentByID(context.Background(), tt.commentID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetCommentByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDeleteComment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCommentRepository(sqlxDB)

	tests := []struct {
		name     string
		commentID int64
		mock     func()
		wantErr  error
	}{
		{
			name: "Success",
			commentID: 1,
			mock: func() {
				mock.ExpectExec(`DELETE FROM comments`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: nil,
		},
		{
			name: "Not Found",
			commentID: 2,
			mock: func() {
				mock.ExpectExec(`DELETE FROM comments`).WithArgs(int64(2)).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: ErrCommentNotFound,
		},
		{
			name: "Database Error",
			commentID: 3,
			mock: func() {
				mock.ExpectExec(`DELETE FROM comments`).WithArgs(int64(3)).WillReturnError(sql.ErrConnDone)
			},
			wantErr: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			err := repo.DeleteComment(context.Background(), tt.commentID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("DeleteComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}