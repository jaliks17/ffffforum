package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"

	pb "github.com/jaliks17/ffffforum/backend/proto"

	"github.com/jaliks17/ffffforum/backend/forum-service/internal/entity"
	"github.com/jaliks17/ffffforum/backend/forum-service/internal/repository"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type MockCommentRepository struct {
	CreateCommentFunc       func(ctx context.Context, comment *entity.Comment) error
	GetCommentsByPostIDFunc func(ctx context.Context, postID int64) ([]entity.Comment, error)
	DeleteCommentFunc       func(ctx context.Context, id int64) error
	GetCommentByIDFunc      func(ctx context.Context, id int64) (*entity.Comment, error)
}

func (m *MockCommentRepository) CreateComment(ctx context.Context, comment *entity.Comment) error {
	return m.CreateCommentFunc(ctx, comment)
}

func (m *MockCommentRepository) GetCommentsByPostID(ctx context.Context, postID int64) ([]entity.Comment, error) {
	return m.GetCommentsByPostIDFunc(ctx, postID)
}

func (m *MockCommentRepository) DeleteComment(ctx context.Context, id int64) error {
	return m.DeleteCommentFunc(ctx, id)
}

func (m *MockCommentRepository) GetCommentByID(ctx context.Context, id int64) (*entity.Comment, error) {
	if m.GetCommentByIDFunc != nil {
		return m.GetCommentByIDFunc(ctx, id)
	}
	return nil, nil
}

func TestCommentUseCase_CreateComment(t *testing.T) {
	tests := []struct {
		name        string
		comment     *entity.Comment
		mockPost    func() *MockPostRepository
		mockComment func() *MockCommentRepository
		mockAuth    func() *MockAuthServiceClient
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Success",
			comment: &entity.Comment{
				PostID:   1,
				Content:  "Test comment",
				AuthorID: 1,
			},
			mockPost: func() *MockPostRepository {
				return &MockPostRepository{
					GetPostByIDFunc: func(ctx context.Context, id int64) (*entity.Post, error) {
						return &entity.Post{ID: id}, nil
					},
				}
			},
			mockComment: func() *MockCommentRepository {
				return &MockCommentRepository{
					CreateCommentFunc: func(ctx context.Context, comment *entity.Comment) error {
						return nil
					},
				}
			},
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{
					GetUserProfileFunc: func(ctx context.Context, in *pb.GetUserProfileRequest, opts ...grpc.CallOption) (*pb.GetUserProfileResponse, error) {
						return &pb.GetUserProfileResponse{
							User: &pb.User{Username: "testuser"},
						}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "Post not found",
			comment: &entity.Comment{
				PostID:   1,
				Content:  "Test comment",
				AuthorID: 1,
			},
			mockPost: func() *MockPostRepository {
				return &MockPostRepository{
					GetPostByIDFunc: func(ctx context.Context, id int64) (*entity.Post, error) {
						return nil, repository.ErrPostNotFound
					},
				}
			},
			mockComment: func() *MockCommentRepository {
				return &MockCommentRepository{}
			},
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{}
			},
			wantErr:     true,
			expectedErr: repository.ErrPostNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPost := tt.mockPost()
			mockComment := tt.mockComment()
			mockAuth := tt.mockAuth()

			uc := NewCommentUseCase(mockComment, mockPost, mockAuth)

			err := uc.CreateComment(context.Background(), tt.comment)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestCommentUseCase_GetCommentsByPostID(t *testing.T) {
	tests := []struct {
		name        string
		postID      int64
		mockComment func() *MockCommentRepository
		mockAuth    func() *MockAuthServiceClient
		want        []entity.Comment
		wantErr     bool
	}{
		{
			name:   "Success with author names",
			postID: 1,
			mockComment: func() *MockCommentRepository {
				return &MockCommentRepository{
					GetCommentsByPostIDFunc: func(ctx context.Context, postID int64) ([]entity.Comment, error) {
						return []entity.Comment{
							{ID: 1, PostID: postID, AuthorID: 1, Content: "Comment 1"},
							{ID: 2, PostID: postID, AuthorID: 2, Content: "Comment 2"},
						}, nil
					},
				}
			},
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{
					GetUserProfileFunc: func(ctx context.Context, in *pb.GetUserProfileRequest, opts ...grpc.CallOption) (*pb.GetUserProfileResponse, error) {
						return &pb.GetUserProfileResponse{
							User: &pb.User{Username: fmt.Sprintf("user%d", in.UserId)},
						}, nil
					},
				}
			},
			want: []entity.Comment{
				{ID: 1, PostID: 1, AuthorID: 1, Content: "Comment 1", AuthorName: "user1"},
				{ID: 2, PostID: 1, AuthorID: 2, Content: "Comment 2", AuthorName: "user2"},
			},
			wantErr: false,
		},
		{
			name:   "Unknown author",
			postID: 1,
			mockComment: func() *MockCommentRepository {
				return &MockCommentRepository{
					GetCommentsByPostIDFunc: func(ctx context.Context, postID int64) ([]entity.Comment, error) {
						return []entity.Comment{
							{ID: 1, PostID: postID, AuthorID: 1, Content: "Comment 1"},
						}, nil
					},
				}
			},
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{
					GetUserProfileFunc: func(ctx context.Context, in *pb.GetUserProfileRequest, opts ...grpc.CallOption) (*pb.GetUserProfileResponse, error) {
						return nil, errors.New("user not found")
					},
				}
			},
			want: []entity.Comment{
				{ID: 1, PostID: 1, AuthorID: 1, Content: "Comment 1", AuthorName: "Unknown"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockComment := tt.mockComment()
			mockPost := &MockPostRepository{}
			mockAuth := tt.mockAuth()

			uc := NewCommentUseCase(mockComment, mockPost, mockAuth)

			got, err := uc.GetCommentsByPostID(context.Background(), tt.postID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCommentsByPostID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCommentUseCase_GetComment(t *testing.T) {
	tests := []struct {
		name string
		commentID int64
		mockComment func() *MockCommentRepository
		want *entity.Comment
		wantErr error
	}{
		{
			name: "Success",
			commentID: 1,
			mockComment: func() *MockCommentRepository {
				return &MockCommentRepository{
					GetCommentByIDFunc: func(ctx context.Context, id int64) (*entity.Comment, error) {
						return &entity.Comment{ID: id, Content: "Test Content"}, nil
					},
				}
			},
			want: &entity.Comment{ID: 1, Content: "Test Content"},
			wantErr: nil,
		},
		{
			name: "Not Found",
			commentID: 2,
			mockComment: func() *MockCommentRepository {
				return &MockCommentRepository{
					GetCommentByIDFunc: func(ctx context.Context, id int64) (*entity.Comment, error) {
						return nil, ErrCommentNotFound
					},
				}
			},
			want: nil,
			wantErr: ErrCommentNotFound,
		},
		{
			name: "Database Error",
			commentID: 3,
			mockComment: func() *MockCommentRepository {
				return &MockCommentRepository{
					GetCommentByIDFunc: func(ctx context.Context, id int64) (*entity.Comment, error) {
						return nil, errors.New("database error")
					},
				}
			},
			want: nil,
			wantErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockComment := tt.mockComment()
			uc := NewCommentUseCase(mockComment, nil, nil)

			got, err := uc.GetComment(context.Background(), tt.commentID)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCommentUseCase_DeleteComment(t *testing.T) {
	tests := []struct {
		name string
		commentID int64
		userID int64
		mockComment func() *MockCommentRepository
		wantErr error
	}{
		{
			name: "Success - Author",
			commentID: 1,
			userID: 1,
			mockComment: func() *MockCommentRepository {
				return &MockCommentRepository{
					GetCommentByIDFunc: func(ctx context.Context, id int64) (*entity.Comment, error) {
						return &entity.Comment{ID: id, AuthorID: 1}, nil
					},
					DeleteCommentFunc: func(ctx context.Context, id int64) error {
						return nil
					},
				}
			},
			wantErr: nil,
		},
		{
			name: "Not Found",
			commentID: 2,
			userID: 1,
			mockComment: func() *MockCommentRepository {
				return &MockCommentRepository{
					GetCommentByIDFunc: func(ctx context.Context, id int64) (*entity.Comment, error) {
						return nil, ErrCommentNotFound
					},
				}
			},
			wantErr: ErrCommentNotFound,
		},
		{
			name: "Forbidden",
			commentID: 1,
			userID: 2,
			mockComment: func() *MockCommentRepository {
				return &MockCommentRepository{
					GetCommentByIDFunc: func(ctx context.Context, id int64) (*entity.Comment, error) {
						return &entity.Comment{ID: id, AuthorID: 1}, nil
					},
				}
			},
			wantErr: ErrForbidden,
		},
		{
			name: "GetCommentByID Error",
			commentID: 1,
			userID: 1,
			mockComment: func() *MockCommentRepository {
				return &MockCommentRepository{
					GetCommentByIDFunc: func(ctx context.Context, id int64) (*entity.Comment, error) {
						return nil, errors.New("database error")
					},
				}
			},
			wantErr: errors.New("database error"),
		},
		{
			name: "DeleteComment Error",
			commentID: 1,
			userID: 1,
			mockComment: func() *MockCommentRepository {
				return &MockCommentRepository{
					GetCommentByIDFunc: func(ctx context.Context, id int64) (*entity.Comment, error) {
						return &entity.Comment{ID: id, AuthorID: 1}, nil
					},
					DeleteCommentFunc: func(ctx context.Context, id int64) error {
						return errors.New("delete error")
					},
				}
			},
			wantErr: errors.New("delete error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockComment := tt.mockComment()
			uc := NewCommentUseCase(mockComment, nil, nil)

			err := uc.DeleteComment(context.Background(), tt.commentID, tt.userID)

			assert.Equal(t, tt.wantErr, err)
		})
	}
}