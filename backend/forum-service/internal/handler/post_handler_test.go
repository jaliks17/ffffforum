package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	pb "backend/proto"
	"forum-service/internal/entity"
	"forum-service/internal/repository"
	"forum-service/internal/usecase"
	"forum-service/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCommentRepository struct {
	mock.Mock
}

func (m *MockCommentRepository) CreateComment(ctx context.Context, comment *entity.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentRepository) GetCommentsByPostID(ctx context.Context, postID int64) ([]entity.Comment, error) {
	args := m.Called(ctx, postID)
	return args.Get(0).([]entity.Comment), args.Error(1)
}

func (m *MockCommentRepository) DeleteComment(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCommentRepository) GetCommentByID(ctx context.Context, id int64) (*entity.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Comment), args.Error(1)
}

type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) GetPostByID(ctx context.Context, id int64) (*entity.Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Post), args.Error(1)
}

func (m *MockPostRepository) CreatePost(ctx context.Context, post *entity.Post) (int64, error) {
	args := m.Called(ctx, post)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPostRepository) GetPosts(ctx context.Context) ([]*entity.Post, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Post), args.Error(1)
}

func (m *MockPostRepository) DeletePost(ctx context.Context, id, authorID int64, role string) error {
	args := m.Called(ctx, id, authorID, role)
	return args.Error(0)
}

func (m *MockPostRepository) UpdatePost(ctx context.Context, id, authorID int64, role, title, content string) (*entity.Post, error) {
	args := m.Called(ctx, id, authorID, role, title, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Post), args.Error(1)
}

type MockPostUsecase struct {
	mock.Mock
}

func (m *MockPostUsecase) CreatePost(ctx context.Context, token, title, content string) (*entity.Post, error) {
	args := m.Called(ctx, token, title, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Post), args.Error(1)
}

func (m *MockPostUsecase) GetPosts(ctx context.Context) ([]*entity.Post, map[int]string, error) {
	args := m.Called(ctx)
	// Handle nil return for posts and authorNames to avoid panic during type assertion
	posts, ok := args.Get(0).([]*entity.Post)
	if !ok && args.Get(0) != nil {
		// This case should not happen if mock is set up correctly, but adding safeguard
		return nil, nil, errors.New("unexpected type for posts")
	}
	authorNames, ok := args.Get(1).(map[int]string)
	if !ok && args.Get(1) != nil {
		// This case should not happen if mock is set up correctly, but adding safeguard
		return nil, nil, errors.New("unexpected type for author names")
	}
	return posts, authorNames, args.Error(2)
}

func (m *MockPostUsecase) DeletePost(ctx context.Context, token string, postID int64) error {
	args := m.Called(ctx, token, postID)
	return args.Error(0)
}

func (m *MockPostUsecase) UpdatePost(ctx context.Context, token string, postID int64, title, content string) (*entity.Post, error) {
	args := m.Called(ctx, token, postID, title, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Post), args.Error(1)
}

func TestCreateComment_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authClient := new(MockAuthClient)
	commentRepo := new(MockCommentRepository)
	postRepo := new(MockPostRepository)

	uc := &usecase.CommentUseCase{
		AuthClient:  authClient,
		CommentRepo: commentRepo,
		PostRepo:    postRepo,
	}

	handler := NewCommentHandler(uc)
	router := gin.Default()
	router.POST("/posts/:id/comments", handler.CreateComment)

	// Mock PostRepo.GetPostByID
	postRepo.On("GetPostByID", mock.Anything, int64(1)).Return(&entity.Post{ID: 1}, nil).Once()

	authClient.On("ValidateToken", mock.Anything, mock.Anything, mock.Anything).Return(&pb.ValidateSessionResponse{Valid: true, UserId: 42}, nil).Once()

	authClient.On("GetUserProfile", mock.Anything, &pb.GetUserProfileRequest{UserId: 42}, mock.Anything).
		Return(&pb.GetUserProfileResponse{User: &pb.User{Username: "alice"}}, nil)

	expectedComment := &entity.Comment{
		Content:    "test comment",
		AuthorID:   42,
		PostID:     1,
		AuthorName: "alice",
	}
	commentRepo.On("CreateComment", mock.Anything, expectedComment).Return(nil)

	body := `{"content":"test comment"}`
	req, _ := http.NewRequest("POST", "/posts/1/comments", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	authClient.AssertExpectations(t)
	commentRepo.AssertExpectations(t)
	postRepo.AssertExpectations(t)
}

func TestGetPosts_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUsecase := new(MockPostUsecase)
	logger, err := logger.NewLogger("info")
	assert.NoError(t, err)
	handler := NewPostHandler(mockUsecase, logger)

	router := gin.Default()
	router.GET("/posts", handler.GetPosts)

	// Mock data
	posts := []*entity.Post{
		{
			ID:        1,
			Title:     "Test Post 1",
			Content:   "Content 1",
			AuthorID:  1,
			CreatedAt: time.Now(),
		},
		{
			ID:        2,
			Title:     "Test Post 2",
			Content:   "Content 2",
			AuthorID:  2,
			CreatedAt: time.Now(),
		},
	}

	authorNames := map[int]string{
		1: "user1",
		2: "user2",
	}

	mockUsecase.On("GetPosts", mock.Anything).Return(posts, authorNames, nil).Once()

	req, _ := http.NewRequest("GET", "/posts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data, ok := response["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 2)

	// Verify first post
	post1 := data[0].(map[string]interface{})
	assert.Equal(t, float64(1), post1["id"])
	assert.Equal(t, "Test Post 1", post1["title"])
	assert.Equal(t, "Content 1", post1["content"])
	assert.Equal(t, float64(1), post1["author_id"])
	assert.Equal(t, "user1", post1["author_name"])

	// Verify second post
	post2 := data[1].(map[string]interface{})
	assert.Equal(t, float64(2), post2["id"])
	assert.Equal(t, "Test Post 2", post2["title"])
	assert.Equal(t, "Content 2", post2["content"])
	assert.Equal(t, float64(2), post2["author_id"])
	assert.Equal(t, "user2", post2["author_name"])

	mockUsecase.AssertExpectations(t)
}

func TestGetPosts_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUsecase := new(MockPostUsecase)
	logger, err := logger.NewLogger("info")
	assert.NoError(t, err)
	handler := NewPostHandler(mockUsecase, logger)

	router := gin.Default()
	router.GET("/posts", handler.GetPosts)

	mockUsecase.On("GetPosts", mock.Anything).Return(nil, nil, errors.New("database error")).Once()

	req, _ := http.NewRequest("GET", "/posts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to get posts", response["error"])
	assert.Equal(t, "database error", response["details"])

	mockUsecase.AssertExpectations(t)
}

func TestDeletePost_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUsecase := new(MockPostUsecase)
	logger, err := logger.NewLogger("info")
	assert.NoError(t, err)
	handler := NewPostHandler(mockUsecase, logger)

	router := gin.Default()
	router.DELETE("/posts/:id", handler.DeletePost)

	mockUsecase.On("DeletePost", mock.Anything, "valid-token", int64(1)).Return(nil).Once()

	req, _ := http.NewRequest("DELETE", "/posts/1", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Post deleted successfully", response["message"])

	mockUsecase.AssertExpectations(t)
}

func TestDeletePost_MissingAuthHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUsecase := new(MockPostUsecase)
	logger, err := logger.NewLogger("info")
	assert.NoError(t, err)
	handler := NewPostHandler(mockUsecase, logger)

	router := gin.Default()
	router.DELETE("/posts/:id", handler.DeletePost)

	req, _ := http.NewRequest("DELETE", "/posts/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Authorization header is required", response["error"])
}

func TestDeletePost_InvalidPostID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUsecase := new(MockPostUsecase)
	logger, err := logger.NewLogger("info")
	assert.NoError(t, err)
	handler := NewPostHandler(mockUsecase, logger)

	router := gin.Default()
	router.DELETE("/posts/:id", handler.DeletePost)

	req, _ := http.NewRequest("DELETE", "/posts/invalid", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid post ID", response["error"])
}

func TestDeletePost_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUsecase := new(MockPostUsecase)
	logger, err := logger.NewLogger("info")
	assert.NoError(t, err)
	handler := NewPostHandler(mockUsecase, logger)

	router := gin.Default()
	router.DELETE("/posts/:id", handler.DeletePost)

	mockUsecase.On("DeletePost", mock.Anything, "valid-token", int64(1)).Return(repository.ErrPostNotFound).Once()

	req, _ := http.NewRequest("DELETE", "/posts/1", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Post not found", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestDeletePost_PermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUsecase := new(MockPostUsecase)
	logger, err := logger.NewLogger("info")
	assert.NoError(t, err)
	handler := NewPostHandler(mockUsecase, logger)

	router := gin.Default()
	router.DELETE("/posts/:id", handler.DeletePost)

	mockUsecase.On("DeletePost", mock.Anything, "valid-token", int64(1)).Return(repository.ErrPermissionDenied).Once()

	req, _ := http.NewRequest("DELETE", "/posts/1", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "You don't have permission", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestUpdatePost_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUsecase := new(MockPostUsecase)
	logger, err := logger.NewLogger("info")
	assert.NoError(t, err)
	handler := NewPostHandler(mockUsecase, logger)

	router := gin.Default()
	router.PUT("/posts/:id", handler.UpdatePost)

	updatedPost := &entity.Post{
		ID:        1,
		Title:     "Updated Title",
		Content:   "Updated Content",
		AuthorID:  42,
		CreatedAt: time.Now(),
	}

	mockUsecase.On("UpdatePost", mock.Anything, "valid-token", int64(1), "Updated Title", "Updated Content").Return(updatedPost, nil).Once()

	body := `{"title":"Updated Title","content":"Updated Content"}`
	req, _ := http.NewRequest("PUT", "/posts/1", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "Post updated successfully", response["message"])
	postData, ok := response["post"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(updatedPost.ID), postData["id"])
	assert.Equal(t, updatedPost.Title, postData["title"])
	assert.Equal(t, updatedPost.Content, postData["content"])

	mockUsecase.AssertExpectations(t)
}

func TestUpdatePost_MissingAuthHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUsecase := new(MockPostUsecase)
	logger, err := logger.NewLogger("info")
	assert.NoError(t, err)
	handler := NewPostHandler(mockUsecase, logger)

	router := gin.Default()
	router.PUT("/posts/:id", handler.UpdatePost)

	body := `{"title":"Updated Title","content":"Updated Content"}`
	req, _ := http.NewRequest("PUT", "/posts/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Authorization header is required", response["error"])
}

func TestUpdatePost_InvalidPostID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUsecase := new(MockPostUsecase)
	logger, err := logger.NewLogger("info")
	assert.NoError(t, err)
	handler := NewPostHandler(mockUsecase, logger)

	router := gin.Default()
	router.PUT("/posts/:id", handler.UpdatePost)

	body := `{"title":"Updated Title","content":"Updated Content"}`
	req, _ := http.NewRequest("PUT", "/posts/invalid", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid post ID", response["error"])
}

func TestUpdatePost_InvalidRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUsecase := new(MockPostUsecase)
	logger, err := logger.NewLogger("info")
	assert.NoError(t, err)
	handler := NewPostHandler(mockUsecase, logger)

	router := gin.Default()
	router.PUT("/posts/:id", handler.UpdatePost)

	body := `{"title":"Updated Title"}` // missing content
	req, _ := http.NewRequest("PUT", "/posts/1", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request body", response["error"])
}

func TestUpdatePost_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUsecase := new(MockPostUsecase)
	logger, err := logger.NewLogger("info")
	assert.NoError(t, err)
	handler := NewPostHandler(mockUsecase, logger)

	router := gin.Default()
	router.PUT("/posts/:id", handler.UpdatePost)

	mockUsecase.On("UpdatePost", mock.Anything, "valid-token", int64(1), "Updated Title", "Updated Content").Return(nil, repository.ErrPostNotFound).Once()

	body := `{"title":"Updated Title","content":"Updated Content"}`
	req, _ := http.NewRequest("PUT", "/posts/1", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Post not found", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestUpdatePost_PermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUsecase := new(MockPostUsecase)
	logger, err := logger.NewLogger("info")
	assert.NoError(t, err)
	handler := NewPostHandler(mockUsecase, logger)

	router := gin.Default()
	router.PUT("/posts/:id", handler.UpdatePost)

	mockUsecase.On("UpdatePost", mock.Anything, "valid-token", int64(1), "Updated Title", "Updated Content").Return(nil, repository.ErrPermissionDenied).Once()

	body := `{"title":"Updated Title","content":"Updated Content"}`
	req, _ := http.NewRequest("PUT", "/posts/1", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Permission denied", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestUpdatePost_UsecaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUsecase := new(MockPostUsecase)
	logger, err := logger.NewLogger("info")
	assert.NoError(t, err)
	handler := NewPostHandler(mockUsecase, logger)

	router := gin.Default()
	router.PUT("/posts/:id", handler.UpdatePost)

	mockUsecase.On("UpdatePost", mock.Anything, "valid-token", int64(1), "Updated Title", "Updated Content").Return(nil, errors.New("some usecase error")).Once()

	body := `{"title":"Updated Title","content":"Updated Content"}`
	req, _ := http.NewRequest("PUT", "/posts/1", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to update post", response["error"])
}

func TestCreatePost_Success(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		reqBody        string
		mockCreateErr  error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "successful creation",
			authHeader: "Bearer valid-token",
			reqBody:    `{"title":"Test Post","content":"This is a test post"}`,
			expectedStatus: 201,
			expectedBody: map[string]interface{}{
				"id":      float64(0),
				"message": "Post created successfully",
				"post": map[string]interface{}{
					"id":         float64(0),
					"title":      "Test Post",
					"content":    "This is a test post",
					"author_id":  float64(0),
					"created_at": "0001-01-01T00:00:00Z",
					"updated_at": "0001-01-01T00:00:00Z",
				},
			},
		},
		{
			name:           "missing auth header",
			authHeader:     "",
			reqBody:        `{"title":"Test Post","content":"This is a test post"}`,
			expectedStatus: 401,
			expectedBody: map[string]interface{}{
				"error": "Authorization header is required",
			},
		},
		{
			name:           "invalid request body",
			authHeader:     "Bearer valid-token",
			reqBody:        `{"title":"Test Post"}`,
			expectedStatus: 400,
			expectedBody: map[string]interface{}{
				"error": "Invalid request body",
			},
		},
		{
			name:           "usecase error",
			authHeader:     "Bearer valid-token",
			reqBody:        `{"title":"Test Post","content":"This is a test post"}`,
			mockCreateErr:  errors.New("failed to create post in usecase"),
			expectedStatus: 500,
			expectedBody: map[string]interface{}{
				"error": "Failed to create post",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("POST", "/posts", bytes.NewBufferString(tt.reqBody))
			c.Request.Header.Set("Content-Type", "application/json")
			if tt.authHeader != "" {
				c.Request.Header.Set("Authorization", tt.authHeader)
			}

			mockUsecase := new(MockPostUsecase)
			logger, _ := logger.NewLogger("info")
			handler := NewPostHandler(mockUsecase, logger)

			// Setup mock expectations for usecase call
			if tt.authHeader != "" && tt.expectedStatus != http.StatusBadRequest && tt.expectedStatus != http.StatusUnauthorized {
				expectedPost := &entity.Post{
					Title:   "Test Post",
					Content: "This is a test post",
				}
				mockUsecase.On("CreatePost", mock.Anything, strings.TrimPrefix(tt.authHeader, "Bearer "), expectedPost.Title, expectedPost.Content).Return(expectedPost, tt.mockCreateErr).Once()
			}

			// Execute
			handler.CreatePost(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			// More robust body assertion
			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}

			mockUsecase.AssertExpectations(t)
		})
	}
}