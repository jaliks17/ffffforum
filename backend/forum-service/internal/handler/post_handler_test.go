package handler

import (
	"context"
	"errors"
	"forum-service/internal/entity"
	"forum-service/pkg/logger"
	"testing"

	pb "backend/proto"

	"github.com/stretchr/testify/assert"
)

type mockPostUseCase struct {
	createErr error
	getErr    error
	deleteErr error
	post      *entity.Post
}

type mockAuthClient struct{}

func (m *mockAuthClient) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateSessionResponse, error) {
	return &pb.ValidateSessionResponse{Valid: true}, nil
}

func (m *mockAuthClient) GetUser(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.UserResponse, error) {
	return &pb.UserResponse{}, nil
}

func (m *mockPostUseCase) CreatePost(ctx context.Context, token, title, content string) (*entity.Post, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return &entity.Post{Title: title, Content: content}, nil
}

func (m *mockPostUseCase) GetPost(ctx context.Context, id int64) (*entity.Post, error) {
	return m.post, m.getErr
}

func (m *mockPostUseCase) DeletePost(ctx context.Context, token string, id int64) error {
	return m.deleteErr
}

func (m *mockPostUseCase) GetPosts(ctx context.Context) ([]*entity.Post, map[int]string, error) {
	return nil, nil, nil
}

func (m *mockPostUseCase) UpdatePost(ctx context.Context, token string, postID int64, title, content string) (*entity.Post, error) {
	return nil, nil
}

func (m *mockPostUseCase) GetPostByID(ctx context.Context, id int64) (*entity.Post, error) {
	return m.post, m.getErr
}

func (m *mockPostUseCase) GetPostsByUserID(ctx context.Context, userID int64) ([]*entity.Post, error) {
	return nil, nil
}

func setupTest(t *testing.T) (*mockPostUseCase, *PostHandler) {
	m := &mockPostUseCase{}
	log, _ := logger.NewLogger("test")
	auth := &mockAuthClient{}
	h := NewPostHandler(m, log, auth)
	return m, h
}

func TestPostHandler_Create(t *testing.T) {
	_, h := setupTest(t)
	err := h.Create(context.Background(), &entity.Post{})
	assert.NoError(t, err)
}

func TestPostHandler_Get(t *testing.T) {
	expected := &entity.Post{ID: 1, Title: "test"}
	m := &mockPostUseCase{post: expected}
	log, _ := logger.NewLogger("test")
	auth := &mockAuthClient{}
	h := NewPostHandler(m, log, auth)
	p, err := h.Get(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, expected, p)
}

func TestPostHandler_Delete(t *testing.T) {
	_, h := setupTest(t)
	err := h.Delete(context.Background(), 1)
	assert.NoError(t, err)
}

func TestPostHandler_Get_Error(t *testing.T) {
	m, h := setupTest(t)
	m.getErr = errors.New("not found")
	_, err := h.Get(context.Background(), 1)
	assert.Error(t, err)
}