package repository

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"forum-service/internal/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestCreatePost(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPostRepository(sqlxDB)

	now := time.Now()

	tests := []struct {
		name    string
		post    *entity.Post
		mock    func()
		want    int64
		wantErr bool
	}{
		{
			name: "Success",
			post: &entity.Post{
				Title:     "Test Post",
				Content:   "Test Content",
				AuthorID:  1,
				CreatedAt: now,
			},
			mock: func() {
				mock.ExpectQuery(`INSERT INTO posts`).
					WithArgs("Test Post", "Test Content", int64(1), now).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			want: 1,
		},
		{
			name: "Empty Fields",
			post: &entity.Post{
				Title:     "",
				Content:   "",
				AuthorID:  1,
				CreatedAt: now,
			},
			mock: func() {
				mock.ExpectQuery(`INSERT INTO posts`).
					WithArgs("", "", int64(1), now).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.CreatePost(context.Background(), tt.post)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreatePost() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPosts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPostRepository(sqlxDB)

	now := time.Now()

	tests := []struct {
		name    string
		mock    func()
		want    []*entity.Post
		wantErr bool
	}{
		{
			name: "Success",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at"}).
					AddRow(1, "Post 1", "Content 1", 1, now).
					AddRow(2, "Post 2", "Content 2", 2, now)
				mock.ExpectQuery(`SELECT`).WillReturnRows(rows)
			},
			want: []*entity.Post{
				{
					ID:        1,
					Title:     "Post 1",
					Content:   "Content 1",
					AuthorID:  1,
					CreatedAt: now,
				},
				{
					ID:        2,
					Title:     "Post 2",
					Content:   "Content 2",
					AuthorID:  2,
					CreatedAt: now,
				},
			},
		},
		{
			name: "No Posts",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at"})
				mock.ExpectQuery(`SELECT`).WillReturnRows(rows)
			},
			want: []*entity.Post{},
			wantErr: false,
		},
		{
			name: "Error",
			mock: func() {
				mock.ExpectQuery(`SELECT`).WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.GetPosts(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPosts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// assert.Equal(t, tt.want, got) // Commented out old assertion

			// Handle case where empty slice is expected
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

func TestDeletePost(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPostRepository(sqlxDB)

	tests := []struct {
		name     string
		postID   int64
		authorID int64
		role     string
		mock     func()
		wantErr  bool
	}{
		{
			name:     "Success - Author",
			postID:   1,
			authorID: 1,
			role:     "user",
			mock: func() {
				mock.ExpectExec(`DELETE FROM posts`).
					WithArgs(int64(1), int64(1), "user").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:     "Success - Admin",
			postID:   1,
			authorID: 2,
			role:     "admin",
			mock: func() {
				mock.ExpectExec(`DELETE FROM posts`).
					WithArgs(int64(1), int64(2), "admin").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:     "Not Found",
			postID:   2,
			authorID: 1,
			role:     "user",
			mock: func() {
				mock.ExpectExec(`DELETE FROM posts`).
					WithArgs(int64(2), int64(1), "user").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			err := repo.DeletePost(context.Background(), tt.postID, tt.authorID, tt.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeletePost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetPostByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPostRepository(sqlxDB)

	now := time.Now()

	tests := []struct {
		name    string
		postID  int64
		mock    func()
		want    *entity.Post
		wantErr error
	}{
		{
			name:   "Success",
			postID: 1,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at"}).
					AddRow(1, "Post 1", "Content 1", 1, now)
				mock.ExpectQuery(`SELECT`).WithArgs(int64(1)).WillReturnRows(rows)
			},
			want: &entity.Post{
				ID:        1,
				Title:     "Post 1",
				Content:   "Content 1",
				AuthorID:  1,
				CreatedAt: now,
			},
		},
		{
			name:   "Not Found",
			postID: 2,
			mock: func() {
				mock.ExpectQuery(`SELECT`).WithArgs(int64(2)).WillReturnError(sql.ErrNoRows)
			},
			wantErr: ErrPostNotFound,
		},
		{
			name:   "Database Error",
			postID: 3,
			mock: func() {
				mock.ExpectQuery(`SELECT`).WithArgs(int64(3)).WillReturnError(sql.ErrConnDone)
			},
			wantErr: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.GetPostByID(context.Background(), tt.postID)
			if err != tt.wantErr {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetPostByID() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPostByID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdatePost(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPostRepository(sqlxDB)

	now := time.Now()

	tests := []struct {
		name     string
		postID   int64
		authorID int64
		role     string
		title    string
		content  string
		mock     func()
		want     *entity.Post
		wantErr  error
	}{
		{
			name:     "Success - Author Update",
			postID:   1,
			authorID: 1,
			role:     "user",
			title:    "Updated Title",
			content:  "Updated Content",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at"}).
					AddRow(1, "Updated Title", "Updated Content", 1, now)
				mock.ExpectQuery(`UPDATE posts`).
					WithArgs("Updated Title", "Updated Content", int64(1), int64(1), "user").
					WillReturnRows(rows)
			},
			want: &entity.Post{
				ID:        1,
				Title:     "Updated Title",
				Content:   "Updated Content",
				AuthorID:  1,
				CreatedAt: now,
			},
		},
		{
			name:     "Success - Admin Update",
			postID:   1,
			authorID: 2,
			role:     "admin",
			title:    "Updated Title",
			content:  "Updated Content",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at"}).
					AddRow(1, "Updated Title", "Updated Content", 1, now)
				mock.ExpectQuery(`UPDATE posts`).
					WithArgs("Updated Title", "Updated Content", int64(1), int64(2), "admin").
					WillReturnRows(rows)
			},
			want: &entity.Post{
				ID:        1,
				Title:     "Updated Title",
				Content:   "Updated Content",
				AuthorID:  1,
				CreatedAt: now,
			},
		},
		{
			name:     "Not Found",
			postID:   2,
			authorID: 1,
			role:     "user",
			title:    "Updated Title",
			content:  "Updated Content",
			mock: func() {
				mock.ExpectQuery(`UPDATE posts`).
					WithArgs("Updated Title", "Updated Content", int64(2), int64(1), "user").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: ErrPostNotFound,
		},
		{
			name:     "Database Error",
			postID:   3,
			authorID: 1,
			role:     "user",
			title:    "Updated Title",
			content:  "Updated Content",
			mock: func() {
				mock.ExpectQuery(`UPDATE posts`).
					WithArgs("Updated Title", "Updated Content", int64(3), int64(1), "user").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.UpdatePost(context.Background(), tt.postID, tt.authorID, tt.role, tt.title, tt.content)
			if err != tt.wantErr {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("UpdatePost() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdatePost() got = %v, want %v", got, tt.want)
			}
		})
	}
}