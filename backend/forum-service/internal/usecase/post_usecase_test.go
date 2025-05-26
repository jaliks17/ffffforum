package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	pb "backend/proto"
	"forum-service/internal/entity"
	"forum-service/internal/repository"
	"forum-service/pkg/logger"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewMockLogger() *logger.Logger {
	return &logger.Logger{
		SugaredLogger: zap.NewNop().Sugar(),
	}
}

func TestPostUsecase_UpdatePost(t *testing.T) {
	now := time.Now()
	updatedPost := &entity.Post{
		ID:        1,
		Title:     "Updated Title",
		Content:   "Updated Content",
		AuthorID:  1,
		CreatedAt: now,
	}

	tests := []struct {
		name        string
		token       string
		postID      int64
		title       string
		content     string
		mockAuth    func() *MockAuthServiceClient
		mockRepo    func() *MockPostRepository
		want        *entity.Post
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "Success - Author Update",
			token:   "valid_token",
			postID:  1,
			title:   "Updated Title",
			content: "Updated Content",
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{
					ValidateTokenFunc: func(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
						return &pb.ValidateSessionResponse{
							Valid:    true,
							UserId:   1,
							UserRole: "user",
						}, nil
					},
				}
			},
			mockRepo: func() *MockPostRepository {
				return &MockPostRepository{
					UpdatePostFunc: func(ctx context.Context, id, authorID int64, role, title, content string) (*entity.Post, error) {
						return updatedPost, nil
					},
				}
			},
			want:    updatedPost,
			wantErr: false,
		},
		{
			name:    "Success - Admin Update",
			token:   "valid_token",
			postID:  1,
			title:   "Updated Title",
			content: "Updated Content",
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{
					ValidateTokenFunc: func(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
						return &pb.ValidateSessionResponse{
							Valid:    true,
							UserId:   2,
							UserRole: "admin",
						}, nil
					},
				}
			},
			mockRepo: func() *MockPostRepository {
				return &MockPostRepository{
					UpdatePostFunc: func(ctx context.Context, id, authorID int64, role, title, content string) (*entity.Post, error) {
						return updatedPost, nil
					},
				}
			},
			want:    updatedPost,
			wantErr: false,
		},
		{
			name:    "Invalid Token",
			token:   "invalid_token",
			postID:  1,
			title:   "Updated Title",
			content: "Updated Content",
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{
					ValidateTokenFunc: func(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
						return &pb.ValidateSessionResponse{
							Valid: false,
						}, nil
					},
				}
			},
			mockRepo: func() *MockPostRepository {
				return &MockPostRepository{
					UpdatePostFunc: func(ctx context.Context, id, authorID int64, role, title, content string) (*entity.Post, error) {
						return nil, nil
					},
				}
			},
			wantErr:     true,
			expectedErr: errors.New("invalid token"),
		},
		{
			name:    "Post Not Found",
			token:   "valid_token",
			postID:  999,
			title:   "Updated Title",
			content: "Updated Content",
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{
					ValidateTokenFunc: func(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
						return &pb.ValidateSessionResponse{
							Valid:    true,
							UserId:   1,
							UserRole: "user",
						}, nil
					},
				}
			},
			mockRepo: func() *MockPostRepository {
				return &MockPostRepository{
					UpdatePostFunc: func(ctx context.Context, id, authorID int64, role, title, content string) (*entity.Post, error) {
						return nil, sql.ErrNoRows
					},
				}
			},
			wantErr:     true,
			expectedErr: sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := tt.mockAuth()
			mockRepo := tt.mockRepo()
			mockLogger := NewMockLogger()

			uc := &PostUsecase{
				postRepo:   mockRepo,
				authClient: mockAuth,
				logger:     mockLogger,
			}

			got, err := uc.UpdatePost(context.Background(), tt.token, tt.postID, tt.title, tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdatePost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostUsecase_DeletePost(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		postID      int64
		mockAuth    func() *MockAuthServiceClient
		mockRepo    func() *MockPostRepository
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "Success - Author Delete",
			token:  "valid_token",
			postID: 1,
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{
					ValidateTokenFunc: func(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
						return &pb.ValidateSessionResponse{
							Valid:    true,
							UserId:   1,
							UserRole: "user",
						}, nil
					},
				}
			},
			mockRepo: func() *MockPostRepository {
				return &MockPostRepository{
					DeletePostFunc: func(ctx context.Context, id, authorID int64, role string) error {
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:   "Success - Admin Delete",
			token:  "valid_token",
			postID: 1,
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{
					ValidateTokenFunc: func(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
						return &pb.ValidateSessionResponse{
							Valid:    true,
							UserId:   2,
							UserRole: "admin",
						}, nil
					},
				}
			},
			mockRepo: func() *MockPostRepository {
				return &MockPostRepository{
					DeletePostFunc: func(ctx context.Context, id, authorID int64, role string) error {
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:   "Invalid Token",
			token:  "invalid_token",
			postID: 1,
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{
					ValidateTokenFunc: func(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
						return &pb.ValidateSessionResponse{
							Valid: false,
						}, nil
					},
				}
			},
			mockRepo: func() *MockPostRepository {
				return &MockPostRepository{
					DeletePostFunc: func(ctx context.Context, id, authorID int64, role string) error {
						return nil
					},
				}
			},
			wantErr:     true,
			expectedErr: errors.New("invalid token"),
		},
		{
			name:   "Post Not Found",
			token:  "valid_token",
			postID: 999,
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{
					ValidateTokenFunc: func(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
						return &pb.ValidateSessionResponse{
							Valid:    true,
							UserId:   1,
							UserRole: "user",
						}, nil
					},
				}
			},
			mockRepo: func() *MockPostRepository {
				return &MockPostRepository{
					DeletePostFunc: func(ctx context.Context, id, authorID int64, role string) error {
						return sql.ErrNoRows
					},
				}
			},
			wantErr:     true,
			expectedErr: errors.New("post not found"),
		},
		{
			name:   "Permission Denied",
			token:  "valid_token",
			postID: 1,
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{
					ValidateTokenFunc: func(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
						return &pb.ValidateSessionResponse{
							Valid:    true,
							UserId:   2,
							UserRole: "user",
						}, nil
					},
				}
			},
			mockRepo: func() *MockPostRepository {
				return &MockPostRepository{
					DeletePostFunc: func(ctx context.Context, id, authorID int64, role string) error {
						return repository.ErrPermissionDenied
					},
				}
			},
			wantErr:     true,
			expectedErr: errors.New("permission denied"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := tt.mockAuth()
			mockRepo := tt.mockRepo()
			mockLogger := NewMockLogger()

			uc := &PostUsecase{
				postRepo:   mockRepo,
				authClient: mockAuth,
				logger:     mockLogger,
			}

			err := uc.DeletePost(context.Background(), tt.token, tt.postID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeletePost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error())
				}
			}
		})
	}
}

func TestPostUsecase_GetPosts(t *testing.T) {
	now := time.Now()
	posts := []*entity.Post{
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
	}

	tests := []struct {
		name        string
		mockAuth    func() *MockAuthServiceClient
		mockRepo    func() *MockPostRepository
		wantPosts   []*entity.Post
		wantNames   map[int]string
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Success with usernames",
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{
					GetUserProfileFunc: func(ctx context.Context, in *pb.GetUserProfileRequest, opts ...grpc.CallOption) (*pb.GetUserProfileResponse, error) {
						return &pb.GetUserProfileResponse{
							User: &pb.User{
								Id:       in.UserId,
								Username: "user" + string(rune(in.UserId)),
							},
						}, nil
					},
				}
			},
			mockRepo: func() *MockPostRepository {
				return &MockPostRepository{
					GetPostsFunc: func(ctx context.Context) ([]*entity.Post, error) {
						return posts, nil
					},
				}
			},
			wantPosts: posts,
			wantNames: map[int]string{
				1: "user1",
				2: "user2",
			},
			wantErr: false,
		},
		{
			name: "Error getting posts",
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{}
			},
			mockRepo: func() *MockPostRepository {
				return &MockPostRepository{
					GetPostsFunc: func(ctx context.Context) ([]*entity.Post, error) {
						return nil, errors.New("database error")
					},
				}
			},
			wantErr:     true,
			expectedErr: errors.New("database error"),
		},
		{
			name: "Partial user info",
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{
					GetUserProfileFunc: func(ctx context.Context, in *pb.GetUserProfileRequest, opts ...grpc.CallOption) (*pb.GetUserProfileResponse, error) {
						if in.UserId == 1 {
							return &pb.GetUserProfileResponse{
								User: &pb.User{Username: "user1"},
							}, nil
						}
						return nil, errors.New("user not found")
					},
				}
			},
			mockRepo: func() *MockPostRepository {
				return &MockPostRepository{
					GetPostsFunc: func(ctx context.Context) ([]*entity.Post, error) {
						return posts, nil
					},
				}
			},
			wantPosts: posts,
			wantNames: map[int]string{
				1: "user1",
				2: "Unknown",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := tt.mockAuth()
			mockRepo := tt.mockRepo()
			mockLogger := NewMockLogger()

			uc := &PostUsecase{
				postRepo:   mockRepo,
				authClient: mockAuth,
				logger:     mockLogger,
			}

			gotPosts, gotNames, err := uc.GetPosts(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPosts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error())
				}
				return
			}

			assert.Equal(t, tt.wantPosts, gotPosts)
			assert.Equal(t, tt.wantNames, gotNames)
		})
	}
}

func TestPostUsecase_CreatePost(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		token       string
		title       string
		content     string
		mockAuth    func() *MockAuthServiceClient
		mockRepo    func() *MockPostRepository
		want        *entity.Post
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "Success",
			token:   "valid_token",
			title:   "Test Title",
			content: "Test Content",
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{
					ValidateTokenFunc: func(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
						return &pb.ValidateSessionResponse{
							Valid:    true,
							UserId:   1,
							UserRole: "user",
						}, nil
					},
				}
			},
			mockRepo: func() *MockPostRepository {
				return &MockPostRepository{
					CreatePostFunc: func(ctx context.Context, post *entity.Post) (int64, error) {
						return 1, nil
					},
				}
			},
			want: &entity.Post{
				ID:        1,
				Title:     "Test Title",
				Content:   "Test Content",
				AuthorID:  1,
				CreatedAt: now,
			},
			wantErr: false,
		},
		{
			name:    "Invalid token",
			token:   "invalid_token",
			title:   "Test Title",
			content: "Test Content",
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{
					ValidateTokenFunc: func(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
						return &pb.ValidateSessionResponse{
							Valid: false,
						}, nil
					},
				}
			},
			mockRepo: func() *MockPostRepository {
				return &MockPostRepository{}
			},
			wantErr:     true,
			expectedErr: errors.New("invalid token"),
		},
		{
			name:    "Create post error",
			token:   "valid_token",
			title:   "Test Title",
			content: "Test Content",
			mockAuth: func() *MockAuthServiceClient {
				return &MockAuthServiceClient{
					ValidateTokenFunc: func(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
						return &pb.ValidateSessionResponse{
							Valid:    true,
							UserId:   1,
							UserRole: "user",
						}, nil
					},
				}
			},
			mockRepo: func() *MockPostRepository {
				return &MockPostRepository{
					CreatePostFunc: func(ctx context.Context, post *entity.Post) (int64, error) {
						return 0, errors.New("create error")
					},
				}
			},
			wantErr:     true,
			expectedErr: errors.New("create error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := tt.mockAuth()
			mockRepo := tt.mockRepo()
			mockLogger := NewMockLogger()

			uc := &PostUsecase{
				postRepo:   mockRepo,
				authClient: mockAuth,
				logger:     mockLogger,
			}

			got, err := uc.CreatePost(context.Background(), tt.token, tt.title, tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error())
				}
				return
			}

			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.Title, got.Title)
			assert.Equal(t, tt.want.Content, got.Content)
			assert.Equal(t, tt.want.AuthorID, got.AuthorID)
			assert.WithinDuration(t, time.Now(), got.CreatedAt, time.Second)
		})
	}
}