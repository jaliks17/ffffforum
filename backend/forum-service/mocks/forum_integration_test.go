package mocks

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pb "github.com/jaliks17/ffffforum/backend/proto"

	"github.com/jaliks17/ffffforum/backend/forum-service/internal/entity"
	"github.com/jaliks17/ffffforum/backend/forum-service/internal/handler"
	"github.com/jaliks17/ffffforum/backend/forum-service/internal/repository"
	"github.com/jaliks17/ffffforum/backend/forum-service/internal/usecase"
	"github.com/jaliks17/ffffforum/backend/forum-service/pkg/logger"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type mockAuthClient struct {
	pb.AuthServiceClient
	validateFunc func(context.Context, *pb.ValidateTokenRequest, ...grpc.CallOption) (*pb.ValidateSessionResponse, error)
	getUserFunc  func(context.Context, *pb.GetUserProfileRequest, ...grpc.CallOption) (*pb.GetUserProfileResponse, error)
}

func (m *mockAuthClient) ValidateToken(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
	return m.validateFunc(ctx, in, opts...)
}

func (m *mockAuthClient) GetUserProfile(ctx context.Context, in *pb.GetUserProfileRequest, opts ...grpc.CallOption) (*pb.GetUserProfileResponse, error) {
	return m.getUserFunc(ctx, in, opts...)
}

type testDependencies struct {
	db          *sql.DB
	mock        sqlmock.Sqlmock
	postRepo    repository.PostRepository
	commentRepo repository.CommentRepository
	postUC      *usecase.PostUsecase
	commentUC   *usecase.CommentUseCase
}

func setupTest(t *testing.T) *testDependencies {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	authClient := &mockAuthClient{
		validateFunc: func(ctx context.Context, req *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
			return &pb.ValidateSessionResponse{
				Valid:    true,
				UserId:   1,
				UserRole: "user",
			}, nil
		},
		getUserFunc: func(ctx context.Context, req *pb.GetUserProfileRequest, opts ...grpc.CallOption) (*pb.GetUserProfileResponse, error) {
			return &pb.GetUserProfileResponse{
				User: &pb.User{
					Id:       req.UserId,
					Username: "testuser",
				},
			}, nil
		},
	}

	postRepo := repository.NewPostRepository(sqlxDB)
	commentRepo := repository.NewCommentRepository(sqlxDB)

	postUC := usecase.NewPostUsecase(postRepo, authClient, nil)
	commentUC := usecase.NewCommentUseCase(commentRepo, postRepo, authClient)

	return &testDependencies{
		db:          db,
		mock:        mock,
		postRepo:    postRepo,
		commentRepo: commentRepo,
		postUC:      postUC,
		commentUC:   commentUC,
	}
}

func TestForumService_Integrated(t *testing.T) {
	t.Run("Success scenarios", func(t *testing.T) {
		deps := setupTest(t)
		defer deps.db.Close()

		t.Run("Create and get post", func(t *testing.T) {
			now := time.Now()
			createQuery := `INSERT INTO posts (title, content, author_id, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
			getQuery := `SELECT id, title, content, author_id, created_at FROM posts WHERE id = $1`

			deps.mock.ExpectQuery(createQuery).
				WithArgs("Test Post", "Test Content", int64(1), sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

			post, err := deps.postUC.CreatePost(context.Background(), "valid_token", "Test Post", "Test Content")
			require.NoError(t, err)
			assert.Equal(t, int64(1), post.ID)

			deps.mock.ExpectQuery(getQuery).
				WithArgs(int64(1)).
				WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at"}).
					AddRow(1, "Test Post", "Test Content", int64(1), now))

			retrievedPost, err := deps.postRepo.GetPostByID(context.Background(), 1)
			require.NoError(t, err)
			assert.Equal(t, post.ID, retrievedPost.ID)
		})

		t.Run("Get posts list", func(t *testing.T) {
			query := `SELECT id, title, content, author_id, created_at FROM posts ORDER BY created_at DESC`
			now := time.Now()

			deps.mock.ExpectQuery(query).
				WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at"}).
					AddRow(1, "First Post", "First Content", int64(1), now).
					AddRow(2, "Second Post", "Second Content", int64(2), now.Add(-time.Hour)))

			posts, authorNames, err := deps.postUC.GetPosts(context.Background())
			require.NoError(t, err)
			assert.Len(t, posts, 2)
			assert.Equal(t, "testuser", authorNames[1])
		})

		t.Run("Create comment", func(t *testing.T) {
			postQuery := `SELECT id, title, content, author_id, created_at FROM posts WHERE id = $1`
			commentQuery := `INSERT INTO comments (content, author_id, post_id, author_name) VALUES ($1, $2, $3, $4) RETURNING id`

			deps.mock.ExpectQuery(postQuery).
				WithArgs(int64(1)).
				WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at"}).
					AddRow(1, "Test Post", "Test Content", int64(1), time.Now()))

			deps.mock.ExpectQuery(commentQuery).
				WithArgs("Test Comment", int64(1), int64(1), "testuser").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

			comment := &entity.Comment{
				Content:  "Test Comment",
				AuthorID: 1,
				PostID:   1,
			}

			err := deps.commentUC.CreateComment(context.Background(), comment)
			require.NoError(t, err)
			assert.Equal(t, int64(1), comment.ID)
		})

		t.Run("Get comments", func(t *testing.T) {
			postQuery := `SELECT id, title, content, author_id, created_at FROM posts WHERE id = $1`
			commentQuery := `SELECT id, content, author_id, post_id, author_name FROM comments WHERE post_id = $1 ORDER BY id DESC`

			deps.mock.ExpectQuery(postQuery).
				WithArgs(int64(1)).
				WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at"}).
					AddRow(1, "Test Post", "Test Content", int64(1), time.Now()))

			deps.mock.ExpectQuery(commentQuery).
				WithArgs(int64(1)).
				WillReturnRows(sqlmock.NewRows([]string{"id", "content", "author_id", "post_id", "author_name"}).
					AddRow(1, "First Comment", int64(1), int64(1), "testuser"))

			comments, err := deps.commentUC.GetCommentsByPostID(context.Background(), 1)
			require.NoError(t, err)
			assert.Len(t, comments, 1)
		})

		t.Run("Update post", func(t *testing.T) {
			query := `UPDATE posts SET title = $1, content = $2 WHERE id = $3 AND (author_id = $4 OR $5 = 'admin') RETURNING id, title, content, author_id, created_at`

			deps.mock.ExpectQuery(query).
				WithArgs("Updated Title", "Updated Content", int64(1), int64(1), "user").
				WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at"}).
					AddRow(1, "Updated Title", "Updated Content", int64(1), time.Now()))

			post, err := deps.postUC.UpdatePost(context.Background(), "valid_token", 1, "Updated Title", "Updated Content")
			require.NoError(t, err)
			assert.Equal(t, "Updated Title", post.Title)
		})

		t.Run("Delete post", func(t *testing.T) {
			query := `DELETE FROM posts WHERE id = $1 AND (author_id = $2 OR $3 = 'admin')`

			deps.mock.ExpectExec(query).
				WithArgs(int64(1), int64(1), "user").
				WillReturnResult(sqlmock.NewResult(0, 1))

			err := deps.postUC.DeletePost(context.Background(), "valid_token", 1)
			require.NoError(t, err)
		})

		require.NoError(t, deps.mock.ExpectationsWereMet())
	})

	t.Run("Error scenarios", func(t *testing.T) {
		deps := setupTest(t)
		defer deps.db.Close()

		t.Run("Create post database error", func(t *testing.T) {
			query := `INSERT INTO posts (title, content, author_id, created_at) VALUES ($1, $2, $3, $4) RETURNING id`

			deps.mock.ExpectQuery(query).
				WithArgs("Bad Post", "Bad Content", int64(1), sqlmock.AnyArg()).
				WillReturnError(errors.New("database error"))

			_, err := deps.postUC.CreatePost(context.Background(), "valid_token", "Bad Post", "Bad Content")
			require.Error(t, err)
		})

		t.Run("Get posts list error", func(t *testing.T) {
			query := `SELECT id, title, content, author_id, created_at FROM posts ORDER BY created_at DESC`

			deps.mock.ExpectQuery(query).
				WillReturnError(errors.New("database error"))

			_, _, err := deps.postUC.GetPosts(context.Background())
			require.Error(t, err)
		})

		t.Run("Create comment for non-existent post", func(t *testing.T) {
			query := `SELECT id, title, content, author_id, created_at FROM posts WHERE id = $1`

			deps.mock.ExpectQuery(query).
				WithArgs(int64(999)).
				WillReturnError(sql.ErrNoRows)

			comment := &entity.Comment{
				Content:  "Test Comment",
				AuthorID: 1,
				PostID:   999,
			}

			err := deps.commentUC.CreateComment(context.Background(), comment)
			require.Error(t, err)
			assert.True(t, errors.Is(err, repository.ErrPostNotFound))
		})

		t.Run("Update non-existent post", func(t *testing.T) {
			query := `UPDATE posts SET title = $1, content = $2 WHERE id = $3 AND (author_id = $4 OR $5 = 'admin') RETURNING id, title, content, author_id, created_at`

			deps.mock.ExpectQuery(query).
				WithArgs("New Title", "New Content", int64(999), int64(1), "user").
				WillReturnError(sql.ErrNoRows)

			_, err := deps.postUC.UpdatePost(context.Background(), "valid_token", 999, "New Title", "New Content")
			require.Error(t, err)
			assert.True(t, errors.Is(err, repository.ErrPostNotFound))
		})

		t.Run("Delete non-existent post", func(t *testing.T) {
			query := `DELETE FROM posts WHERE id = $1 AND (author_id = $2 OR $3 = 'admin')`

			deps.mock.ExpectExec(query).
				WithArgs(int64(999), int64(1), "user").
				WillReturnResult(sqlmock.NewResult(0, 0))

			err := deps.postUC.DeletePost(context.Background(), "valid_token", 999)
			require.Error(t, err)
			assert.True(t, errors.Is(err, repository.ErrPostNotFound))
		})

		t.Run("Authentication error", func(t *testing.T) {
			errorAuthClient := &mockAuthClient{
				validateFunc: func(ctx context.Context, req *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
					return nil, errors.New("auth service error")
				},
			}

			errorPostUC := usecase.NewPostUsecase(deps.postRepo, errorAuthClient, nil)

			_, err := errorPostUC.CreatePost(context.Background(), "invalid_token", "Test", "Content")
			require.Error(t, err)
		})

		require.NoError(t, deps.mock.ExpectationsWereMet())
	})

	t.Run("Edge cases", func(t *testing.T) {
		deps := setupTest(t)
		defer deps.db.Close()

		t.Run("Empty posts list", func(t *testing.T) {
			query := `SELECT id, title, content, author_id, created_at FROM posts ORDER BY created_at DESC`

			deps.mock.ExpectQuery(query).
				WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at"}))

			posts, authorNames, err := deps.postUC.GetPosts(context.Background())
			require.NoError(t, err)
			assert.Empty(t, posts)
			assert.Empty(t, authorNames)
		})

		t.Run("Create comment database error", func(t *testing.T) {
			postQuery := `SELECT id, title, content, author_id, created_at FROM posts WHERE id = $1`
			commentQuery := `INSERT INTO comments (content, author_id, post_id, author_name) VALUES ($1, $2, $3, $4) RETURNING id`

			deps.mock.ExpectQuery(postQuery).
				WithArgs(int64(1)).
				WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at"}).
					AddRow(1, "Test Post", "Test Content", int64(1), time.Now()))

			deps.mock.ExpectQuery(commentQuery).
				WithArgs("Bad Comment", int64(1), int64(1), "testuser").
				WillReturnError(errors.New("database error"))

			comment := &entity.Comment{
				Content:  "Bad Comment",
				AuthorID: 1,
				PostID:   1,
			}

			err := deps.commentUC.CreateComment(context.Background(), comment)
			require.Error(t, err)
		})

		t.Run("Update post as admin", func(t *testing.T) {

			authClient := &mockAuthClient{
				validateFunc: func(ctx context.Context, req *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
					return &pb.ValidateSessionResponse{
						Valid:    true,
						UserId:   2,
						UserRole: "admin",
					}, nil
				},
			}

			postUC := usecase.NewPostUsecase(deps.postRepo, authClient, nil)

			query := `UPDATE posts SET title = $1, content = $2 WHERE id = $3 AND (author_id = $4 OR $5 = 'admin') RETURNING id, title, content, author_id, created_at`

			deps.mock.ExpectQuery(query).
				WithArgs("Admin Updated", "Admin Content", int64(1), int64(2), "admin").
				WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at"}).
					AddRow(1, "Admin Updated", "Admin Content", int64(1), time.Now()))

			_, err := postUC.UpdatePost(context.Background(), "admin_token", 1, "Admin Updated", "Admin Content")
			require.NoError(t, err)
		})

		t.Run("Get user error", func(t *testing.T) {
			authClient := &mockAuthClient{
				validateFunc: func(ctx context.Context, req *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
					return &pb.ValidateSessionResponse{
						Valid:    true,
						UserId:   1,
						UserRole: "user",
					}, nil
				},
				getUserFunc: func(ctx context.Context, req *pb.GetUserProfileRequest, opts ...grpc.CallOption) (*pb.GetUserProfileResponse, error) {
					return nil, errors.New("user service error")
				},
			}

			commentUC := usecase.NewCommentUseCase(deps.commentRepo, deps.postRepo, authClient)

			deps.mock.ExpectQuery(`SELECT id, title, content, author_id, created_at FROM posts WHERE id = $1`).
				WithArgs(int64(1)).
				WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at"}).
					AddRow(1, "Test Post", "Test Content", int64(1), time.Now()))

			comment := &entity.Comment{
				Content:  "Test Comment",
				AuthorID: 1,
				PostID:   1,
			}

			err := commentUC.CreateComment(context.Background(), comment)
			require.Error(t, err)
		})

		t.Run("Invalid token", func(t *testing.T) {
			authClient := &mockAuthClient{
				validateFunc: func(ctx context.Context, req *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
					return &pb.ValidateSessionResponse{
						Valid: false,
					}, nil
				},
			}

			postUC := usecase.NewPostUsecase(deps.postRepo, authClient, nil)

			_, err := postUC.CreatePost(context.Background(), "invalid_token", "Test", "Content")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid token")
		})

		t.Run("Update post without permission", func(t *testing.T) {

			authClient := &mockAuthClient{
				validateFunc: func(ctx context.Context, req *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
					return &pb.ValidateSessionResponse{
						Valid:    true,
						UserId:   2,
						UserRole: "user",
					}, nil
				},
			}

			postUC := usecase.NewPostUsecase(deps.postRepo, authClient, nil)

			query := `UPDATE posts SET title = $1, content = $2 WHERE id = $3 AND (author_id = $4 OR $5 = 'admin') RETURNING id, title, content, author_id, created_at`

			deps.mock.ExpectQuery(query).
				WithArgs("New Title", "New Content", int64(1), int64(2), "user").
				WillReturnError(repository.ErrPermissionDenied)

			_, err := postUC.UpdatePost(context.Background(), "valid_token", 1, "New Title", "New Content")
			require.Error(t, err)
			assert.True(t, errors.Is(err, repository.ErrPermissionDenied))
		})
		t.Run("Get comments database error", func(t *testing.T) {
			postQuery := `SELECT id, title, content, author_id, created_at FROM posts WHERE id = $1`
			commentQuery := `SELECT id, content, author_id, post_id, author_name FROM comments WHERE post_id = $1 ORDER BY id DESC`

			deps.mock.ExpectQuery(postQuery).
				WithArgs(int64(1)).
				WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at"}).
					AddRow(1, "Test Post", "Test Content", int64(1), time.Now()))

			deps.mock.ExpectQuery(commentQuery).
				WithArgs(int64(1)).
				WillReturnError(errors.New("database error"))

			_, err := deps.commentUC.GetCommentsByPostID(context.Background(), 1)
			require.Error(t, err)
		})

		t.Run("Empty comments list", func(t *testing.T) {
			postQuery := `SELECT id, title, content, author_id, created_at FROM posts WHERE id = $1`
			commentQuery := `SELECT id, content, author_id, post_id, author_name FROM comments WHERE post_id = $1 ORDER BY id DESC`

			deps.mock.ExpectQuery(postQuery).
				WithArgs(int64(1)).
				WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at"}).
					AddRow(1, "Test Post", "Test Content", int64(1), time.Now()))

			deps.mock.ExpectQuery(commentQuery).
				WithArgs(int64(1)).
				WillReturnRows(sqlmock.NewRows([]string{"id", "content", "author_id", "post_id", "author_name"}))

			comments, err := deps.commentUC.GetCommentsByPostID(context.Background(), 1)
			require.NoError(t, err)
			assert.Empty(t, comments)
		})

		require.NoError(t, deps.mock.ExpectationsWereMet())
	})

}
func TestPostHandler(t *testing.T) {

	mockLogger := &logger.Logger{}

	t.Run("CreatePost success", func(t *testing.T) {
		mockUC := &mockPostUseCase{
			createFunc: func(ctx context.Context, token, title, content string) (*entity.Post, error) {
				return &entity.Post{
					ID:        1,
					Title:     title,
					Content:   content,
					AuthorID:  1,
					CreatedAt: time.Now(),
				}, nil
			},
		}

		handler := handler.NewPostHandler(mockUC, mockLogger)

		router := gin.Default()
		router.POST("/posts", handler.CreatePost)

		requestBody := `{"title": "Test Post", "content": "Test Content"}`
		req, _ := http.NewRequest("POST", "/posts", bytes.NewBufferString(requestBody))
		req.Header.Set("Authorization", "Bearer valid_token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
	t.Run("DeletePost success", func(t *testing.T) {
		mockUC := &mockPostUseCase{
			deleteFunc: func(ctx context.Context, token string, postID int64) error {
				return nil
			},
		}

		handler := handler.NewPostHandler(mockUC, mockLogger)

		router := gin.Default()
		router.DELETE("/posts/:id", handler.DeletePost)

		req, _ := http.NewRequest("DELETE", "/posts/1", nil)
		req.Header.Set("Authorization", "Bearer valid_token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("DeletePost not found", func(t *testing.T) {
		mockUC := &mockPostUseCase{
			deleteFunc: func(ctx context.Context, token string, postID int64) error {
				return repository.ErrPostNotFound
			},
		}

		handler := handler.NewPostHandler(mockUC, mockLogger)

		router := gin.Default()
		router.DELETE("/posts/:id", handler.DeletePost)

		req, _ := http.NewRequest("DELETE", "/posts/999", nil)
		req.Header.Set("Authorization", "Bearer valid_token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("UpdatePost success", func(t *testing.T) {
		mockUC := &mockPostUseCase{
			updateFunc: func(ctx context.Context, token string, postID int64, title, content string) (*entity.Post, error) {
				return &entity.Post{
					ID:        postID,
					Title:     title,
					Content:   content,
					AuthorID:  1,
					CreatedAt: time.Now(),
				}, nil
			},
		}

		handler := handler.NewPostHandler(mockUC, mockLogger)

		router := gin.Default()
		router.PUT("/posts/:id", handler.UpdatePost)

		requestBody := `{"title": "Updated Title", "content": "Updated Content"}`
		req, _ := http.NewRequest("PUT", "/posts/1", bytes.NewBufferString(requestBody))
		req.Header.Set("Authorization", "Bearer valid_token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("CreatePost missing auth header", func(t *testing.T) {
		handler := handler.NewPostHandler(nil, mockLogger)

		router := gin.Default()
		router.POST("/posts", handler.CreatePost)

		requestBody := `{"title": "Test Post", "content": "Test Content"}`
		req, _ := http.NewRequest("POST", "/posts", bytes.NewBufferString(requestBody))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("CreatePost invalid request body", func(t *testing.T) {
		handler := handler.NewPostHandler(nil, mockLogger)

		router := gin.Default()
		router.POST("/posts", handler.CreatePost)

		requestBody := `{"invalid": "data"}`
		req, _ := http.NewRequest("POST", "/posts", bytes.NewBufferString(requestBody))
		req.Header.Set("Authorization", "Bearer valid_token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GetPosts database error", func(t *testing.T) {
		mockUC := &mockPostUseCase{
			getPostsFunc: func(ctx context.Context) ([]*entity.Post, map[int]string, error) {
				return nil, nil, errors.New("database error")
			},
		}

		handler := handler.NewPostHandler(mockUC, mockLogger)

		router := gin.Default()
		router.GET("/posts", handler.GetPosts)

		req, _ := http.NewRequest("GET", "/posts", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
	t.Run("DeletePost permission denied", func(t *testing.T) {
		mockUC := &mockPostUseCase{
			deleteFunc: func(ctx context.Context, token string, postID int64) error {
				return repository.ErrPermissionDenied
			},
		}

		handler := handler.NewPostHandler(mockUC, mockLogger)

		router := gin.Default()
		router.DELETE("/posts/:id", handler.DeletePost)

		req, _ := http.NewRequest("DELETE", "/posts/1", nil)
		req.Header.Set("Authorization", "Bearer valid_token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("DeletePost database error", func(t *testing.T) {
		mockUC := &mockPostUseCase{
			deleteFunc: func(ctx context.Context, token string, postID int64) error {
				return errors.New("database error")
			},
		}

		handler := handler.NewPostHandler(mockUC, mockLogger)

		router := gin.Default()
		router.DELETE("/posts/:id", handler.DeletePost)

		req, _ := http.NewRequest("DELETE", "/posts/1", nil)
		req.Header.Set("Authorization", "Bearer valid_token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("UpdatePost permission denied", func(t *testing.T) {
		mockUC := &mockPostUseCase{
			updateFunc: func(ctx context.Context, token string, postID int64, title, content string) (*entity.Post, error) {
				return nil, repository.ErrPermissionDenied
			},
		}

		handler := handler.NewPostHandler(mockUC, mockLogger)

		router := gin.Default()
		router.PUT("/posts/:id", handler.UpdatePost)

		requestBody := `{"title": "Updated Title", "content": "Updated Content"}`
		req, _ := http.NewRequest("PUT", "/posts/1", bytes.NewBufferString(requestBody))
		req.Header.Set("Authorization", "Bearer valid_token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("UpdatePost database error", func(t *testing.T) {
		mockUC := &mockPostUseCase{
			updateFunc: func(ctx context.Context, token string, postID int64, title, content string) (*entity.Post, error) {
				return nil, errors.New("database error")
			},
		}

		handler := handler.NewPostHandler(mockUC, mockLogger)

		router := gin.Default()
		router.PUT("/posts/:id", handler.UpdatePost)

		requestBody := `{"title": "Updated Title", "content": "Updated Content"}`
		req, _ := http.NewRequest("PUT", "/posts/1", bytes.NewBufferString(requestBody))
		req.Header.Set("Authorization", "Bearer valid_token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
	t.Run("CreatePost invalid token", func(t *testing.T) {
		mockAuth := &mockAuthClient{
			validateFunc: func(ctx context.Context, req *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
				return &pb.ValidateSessionResponse{
					Valid: false,
				}, nil
			},
		}

		postUC := usecase.NewPostUsecase(nil, mockAuth, nil)
		handler := handler.NewPostHandler(postUC, mockLogger)

		router := gin.Default()
		router.POST("/posts", handler.CreatePost)

		requestBody := `{"title": "Test Post", "content": "Test Content"}`
		req, _ := http.NewRequest("POST", "/posts", bytes.NewBufferString(requestBody))
		req.Header.Set("Authorization", "Bearer invalid_token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("UpdatePost invalid request body", func(t *testing.T) {
		handler := handler.NewPostHandler(nil, mockLogger)

		router := gin.Default()
		router.PUT("/posts/:id", handler.UpdatePost)

		requestBody := `{"invalid": "data"}`
		req, _ := http.NewRequest("PUT", "/posts/1", bytes.NewBufferString(requestBody))
		req.Header.Set("Authorization", "Bearer valid_token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("DeletePost invalid token", func(t *testing.T) {
		mockAuth := &mockAuthClient{
			validateFunc: func(ctx context.Context, req *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
				return &pb.ValidateSessionResponse{
					Valid: false,
				}, nil
			},
		}

		postUC := usecase.NewPostUsecase(nil, mockAuth, nil)
		handler := handler.NewPostHandler(postUC, mockLogger)

		router := gin.Default()
		router.DELETE("/posts/:id", handler.DeletePost)

		req, _ := http.NewRequest("DELETE", "/posts/1", nil)
		req.Header.Set("Authorization", "Bearer invalid_token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("UpdatePost invalid post id", func(t *testing.T) {
		handler := handler.NewPostHandler(nil, mockLogger)

		router := gin.Default()
		router.PUT("/posts/:id", handler.UpdatePost)

		requestBody := `{"title": "Updated Title", "content": "Updated Content"}`
		req, _ := http.NewRequest("PUT", "/posts/invalid", bytes.NewBufferString(requestBody))
		req.Header.Set("Authorization", "Bearer valid_token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("GetPosts success", func(t *testing.T) {
		mockUC := &mockPostUseCase{
			getPostsFunc: func(ctx context.Context) ([]*entity.Post, map[int]string, error) {
				return []*entity.Post{
					{
						ID:        1,
						Title:     "Test Post",
						Content:   "Test Content",
						AuthorID:  1,
						CreatedAt: time.Now(),
					},
				}, map[int]string{1: "testuser"}, nil
			},
		}

		handler := handler.NewPostHandler(mockUC, mockLogger)

		router := gin.Default()
		router.GET("/posts", handler.GetPosts)

		req, _ := http.NewRequest("GET", "/posts", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestCommentHandler(t *testing.T) {
	t.Run("CreateComment success", func(t *testing.T) {
		mockAuth := &mockAuthClient{
			validateFunc: func(ctx context.Context, req *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
				return &pb.ValidateSessionResponse{
					Valid:    true,
					UserId:   1,
					UserRole: "user",
				}, nil
			},
			getUserFunc: func(ctx context.Context, req *pb.GetUserProfileRequest, opts ...grpc.CallOption) (*pb.GetUserProfileResponse, error) {
				return &pb.GetUserProfileResponse{
					User: &pb.User{
						Username: "testuser",
					},
				}, nil
			},
		}

		commentUC := usecase.NewCommentUseCase(nil, nil, mockAuth)
		handler := handler.NewCommentHandler(commentUC)

		router := gin.Default()
		router.POST("/posts/:id/comments", handler.CreateComment)

		requestBody := `{"content": "Test Comment"}`
		req, _ := http.NewRequest("POST", "/posts/1/comments", bytes.NewBufferString(requestBody))
		req.Header.Set("Authorization", "Bearer valid_token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "testuser", response["author_name"])
	})
	t.Run("CreateComment invalid post id", func(t *testing.T) {
		handler := handler.NewCommentHandler(nil)

		router := gin.Default()
		router.POST("/posts/:id/comments", handler.CreateComment)

		requestBody := `{"content": "Test Comment"}`
		req, _ := http.NewRequest("POST", "/posts/invalid/comments", bytes.NewBufferString(requestBody))
		req.Header.Set("Authorization", "Bearer valid_token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateComment invalid request body", func(t *testing.T) {
		handler := handler.NewCommentHandler(nil)

		router := gin.Default()
		router.POST("/posts/:id/comments", handler.CreateComment)

		requestBody := `{"invalid": "data"}`
		req, _ := http.NewRequest("POST", "/posts/1/comments", bytes.NewBufferString(requestBody))
		req.Header.Set("Authorization", "Bearer valid_token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GetCommentsByPostID invalid post id", func(t *testing.T) {
		handler := handler.NewCommentHandler(nil)

		router := gin.Default()
		router.GET("/posts/:id/comments", handler.GetCommentsByPostID)

		req, _ := http.NewRequest("GET", "/posts/invalid/comments", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateComment database error", func(t *testing.T) {
		mockAuth := &mockAuthClient{
			validateFunc: func(ctx context.Context, req *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
				return &pb.ValidateSessionResponse{
					Valid:    true,
					UserId:   1,
					UserRole: "user",
				}, nil
			},
			getUserFunc: func(ctx context.Context, req *pb.GetUserProfileRequest, opts ...grpc.CallOption) (*pb.GetUserProfileResponse, error) {
				return &pb.GetUserProfileResponse{
					User: &pb.User{
						Username: "testuser",
					},
				}, nil
			},
		}

		mockUC := &mockCommentUseCase{
			createCommentFunc: func(ctx context.Context, comment *entity.Comment) error {
				return errors.New("database error")
			},
		}

		commentUC := usecase.NewCommentUseCase(mockUC, nil, mockAuth)
		handler := handler.NewCommentHandler(commentUC)

		router := gin.Default()
		router.POST("/posts/:id/comments", handler.CreateComment)

		requestBody := `{"content": "Test Comment"}`
		req, _ := http.NewRequest("POST", "/posts/1/comments", bytes.NewBufferString(requestBody))
		req.Header.Set("Authorization", "Bearer valid_token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("GetCommentsByPostID database error", func(t *testing.T) {

		deps := setupTest(t)
		defer deps.db.Close()

		postQuery := `SELECT id, title, content, author_id, created_at FROM posts WHERE id = $1`
		commentQuery := `SELECT id, content, author_id, post_id, author_name FROM comments WHERE post_id = $1 ORDER BY id DESC`

		deps.mock.ExpectQuery(postQuery).
			WithArgs(int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at"}).
				AddRow(1, "Test Post", "Test Content", int64(1), time.Now()))

		deps.mock.ExpectQuery(commentQuery).
			WithArgs(int64(1)).
			WillReturnError(errors.New("database error"))

		handler := handler.NewCommentHandler(deps.commentUC)

		router := gin.Default()
		router.GET("/posts/:id/comments", handler.GetCommentsByPostID)

		req, _ := http.NewRequest("GET", "/posts/1/comments", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("GetCommentsByPostID success", func(t *testing.T) {

		deps := setupTest(t)
		defer deps.db.Close()

		postQuery := `SELECT id, title, content, author_id, created_at FROM posts WHERE id = $1`
		commentQuery := `SELECT id, content, author_id, post_id, author_name FROM comments WHERE post_id = $1 ORDER BY id DESC`

		deps.mock.ExpectQuery(postQuery).
			WithArgs(int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at"}).
				AddRow(1, "Test Post", "Test Content", int64(1), time.Now()))

		deps.mock.ExpectQuery(commentQuery).
			WithArgs(int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "content", "author_id", "post_id", "author_name"}).
				AddRow(1, "Test Comment", int64(1), int64(1), "testuser"))

		handler := handler.NewCommentHandler(deps.commentUC)

		router := gin.Default()
		router.GET("/posts/:id/comments", handler.GetCommentsByPostID)

		req, _ := http.NewRequest("GET", "/posts/1/comments", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		comments := response["comments"].([]interface{})
		assert.Len(t, comments, 1)
	})
}

type mockPostUseCase struct {
	usecase.PostUsecaseInterface
	createFunc   func(context.Context, string, string, string) (*entity.Post, error)
	getPostsFunc func(context.Context) ([]*entity.Post, map[int]string, error)
	deleteFunc   func(context.Context, string, int64) error
	updateFunc   func(context.Context, string, int64, string, string) (*entity.Post, error)
}

func (m *mockPostUseCase) CreatePost(ctx context.Context, token, title, content string) (*entity.Post, error) {
	return m.createFunc(ctx, token, title, content)
}

func (m *mockPostUseCase) GetPosts(ctx context.Context) ([]*entity.Post, map[int]string, error) {
	return m.getPostsFunc(ctx)
}

func (m *mockPostUseCase) DeletePost(ctx context.Context, token string, postID int64) error {
	return m.deleteFunc(ctx, token, postID)
}

func (m *mockPostUseCase) UpdatePost(ctx context.Context, token string, postID int64, title, content string) (*entity.Post, error) {
	return m.updateFunc(ctx, token, postID, title, content)
}

type mockCommentUseCase struct {
	usecase.CommentUseCase
	createCommentFunc   func(ctx context.Context, comment *entity.Comment) error
	getCommentsFunc     func(ctx context.Context, postID int64) ([]entity.Comment, error)
	deleteCommentFunc   func(ctx context.Context, id int64) error
	getCommentByIDFunc  func(ctx context.Context, id int64) (*entity.Comment, error)
}

func (m *mockCommentUseCase) CreateComment(ctx context.Context, comment *entity.Comment) error {
	if m.createCommentFunc != nil {
		return m.createCommentFunc(ctx, comment)
	}
	return nil
}

func (m *mockCommentUseCase) GetCommentsByPostID(ctx context.Context, postID int64) ([]entity.Comment, error) {
	if m.getCommentsFunc != nil {
		return m.getCommentsFunc(ctx, postID)
	}
	return nil, nil
}

func (m *mockCommentUseCase) DeleteComment(ctx context.Context, commentID int64) error {
	if m.deleteCommentFunc != nil {
		return m.deleteCommentFunc(ctx, commentID)
	}
	return nil
}

func (m *mockCommentUseCase) GetCommentByID(ctx context.Context, id int64) (*entity.Comment, error) {
	if m.getCommentByIDFunc != nil {
		return m.getCommentByIDFunc(ctx, id)
	}
	return nil, nil
}