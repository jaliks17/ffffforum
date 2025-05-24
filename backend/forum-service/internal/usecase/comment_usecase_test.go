package usecase

import (
	"context"
	"errors"
	"forum-service/internal/entity"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockCommentRepo struct {
	createErr error
	getErr    error
	deleteErr error
	comment   *entity.Comment
}

type mockPostRepo struct{}

func (m *mockPostRepo) GetPostByID(ctx context.Context, id int64) (*entity.Post, error) {
	return &entity.Post{}, nil
}

func (m *mockCommentRepo) Create(ctx context.Context, comment *entity.Comment) error {
	return m.createErr
}

func (m *mockCommentRepo) GetByID(ctx context.Context, id int64) (*entity.Comment, error) {
	return m.comment, m.getErr
}

func (m *mockCommentRepo) Delete(ctx context.Context, id int64) error {
	return m.deleteErr
}

func (m *mockCommentRepo) DeleteWithUserID(ctx context.Context, id int64, userID int64) error {
	return m.deleteErr
}

func (m *mockCommentRepo) GetCommentsByPostID(ctx context.Context, postID int64) ([]entity.Comment, error) {
	return []entity.Comment{}, nil
}

func TestCommentUseCase_CreateComment(t *testing.T) {
	mockComment := &mockCommentRepo{}
	mockPost := &mockPostRepo{}
	uc := NewCommentUseCase(mockComment, mockPost)
	err := uc.CreateComment(context.Background(), &entity.Comment{})
	assert.NoError(t, err)
}

func TestCommentUseCase_GetComment(t *testing.T) {
	expected := &entity.Comment{ID: 1, Content: "test"}
	mockComment := &mockCommentRepo{comment: expected}
	mockPost := &mockPostRepo{}
	uc := NewCommentUseCase(mockComment, mockPost)
	c, err := uc.GetComment(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, expected, c)
}

func TestCommentUseCase_DeleteComment(t *testing.T) {
	mockComment := &mockCommentRepo{}
	mockPost := &mockPostRepo{}
	uc := NewCommentUseCase(mockComment, mockPost)
	err := uc.DeleteComment(context.Background(), 1, 1)
	assert.NoError(t, err)
}

func TestCommentUseCase_GetComment_Error(t *testing.T) {
	mockComment := &mockCommentRepo{getErr: errors.New("not found")}
	mockPost := &mockPostRepo{}
	uc := NewCommentUseCase(mockComment, mockPost)
	_, err := uc.GetComment(context.Background(), 1)
	assert.Error(t, err)
}