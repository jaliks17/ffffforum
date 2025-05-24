package usecase_test

import (
	"context"
	"errors"
	"testing"

	pb "backend/proto"
	"forum-service/internal/entity"
	"forum-service/internal/usecase"
	"forum-service/pkg/logger"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestPostUsecase_CreatePost(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	// Create mocks
	mockPostRepo := &MockPostRepository{}
	mockAuthClient := &MockAuthServiceClient{}
	mockLogger, err := logger.NewLogger("development")
	assert.NoError(err, "Failed to create logger")
	if err != nil {
		return
	}

	// Create usecase instance
	postUsecase := usecase.NewPostUsecase(mockPostRepo, mockAuthClient, mockLogger)

	// Define test data
	token := "valid_token"
	title := "Test Post Title"
	content := "Test Post Content"
	userID := int64(1)

	// Set up mock expectations for successful case
	mockAuthClient.ValidateTokenFunc = func(ctx context.Context, req *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
		return &pb.ValidateSessionResponse{Valid: true, UserId: userID, UserRole: "user"}, nil
	}
	mockPostRepo.CreatePostFunc = func(ctx context.Context, post *entity.Post) (int64, error) {
		return 1, nil
	}

	// Call the method
	post, err := postUsecase.CreatePost(ctx, token, title, content)

	// Assertions
	assert.NoError(err)
	assert.NotNil(post)
	assert.Equal(int64(1), post.ID)
	assert.Equal(title, post.Title)
	assert.Equal(content, post.Content)
	assert.Equal(userID, post.AuthorID)

	// --- Test case: Invalid token ---
	token = "invalid_token"

	// Set up mock expectations
	mockAuthClient.ValidateTokenFunc = func(ctx context.Context, req *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
		return &pb.ValidateSessionResponse{Valid: false}, nil
	}

	// Call the method
	post, err = postUsecase.CreatePost(ctx, token, title, content)

	// Assertions
	assert.Error(err)
	assert.Nil(post)
	assert.EqualError(err, "invalid token")

	// --- Test case: Auth service error ---
	token = "token_with_auth_error"
	authError := errors.New("auth service error")

	// Set up mock expectations
	mockAuthClient.ValidateTokenFunc = func(ctx context.Context, req *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
		return nil, authError
	}

	// Call the method
	post, err = postUsecase.CreatePost(ctx, token, title, content)

	// Assertions
	assert.Error(err)
	assert.Nil(post)
	assert.EqualError(err, authError.Error())

	// --- Test case: Repository error ---
	token = "token_with_repo_error"
	repoError := errors.New("repository error")

	// Set up mock expectations
	mockAuthClient.ValidateTokenFunc = func(ctx context.Context, req *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
		return &pb.ValidateSessionResponse{Valid: true, UserId: userID, UserRole: "user"}, nil
	}
	mockPostRepo.CreatePostFunc = func(ctx context.Context, post *entity.Post) (int64, error) {
		return 0, repoError
	}

	// Call the method
	post, err = postUsecase.CreatePost(ctx, token, title, content)

	// Assertions
	assert.Error(err)
	assert.Nil(post)
	assert.EqualError(err, repoError.Error())
} 