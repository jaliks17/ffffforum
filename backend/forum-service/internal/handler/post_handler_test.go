package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "backend/proto"
	"forum-service/internal/entity"
	"forum-service/internal/usecase"

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

func TestCreateComment_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authClient := new(MockAuthClient)
	commentRepo := new(MockCommentRepository)

	uc := &usecase.CommentUseCase{
		AuthClient:  authClient,
		CommentRepo: commentRepo,
	}

	handler := NewCommentHandler(uc)
	router := gin.Default()
	router.POST("/posts/:id/comments", handler.CreateComment)

	authClient.On("ValidateToken", mock.Anything, &pb.ValidateTokenRequest{Token: "valid-token"}, mock.Anything).
		Return(&pb.ValidateSessionResponse{Valid: true, UserId: 42}, nil)

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
}