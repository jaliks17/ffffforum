package usecase_test

import (
	"context"

	"google.golang.org/grpc"

	pb "backend/proto"
	"forum-service/internal/entity"
)

// Mock implementation of PostRepository interface
type MockPostRepository struct {
	CreatePostFunc      func(ctx context.Context, post *entity.Post) (int64, error)
	GetPostsFunc        func(ctx context.Context) ([]*entity.Post, error)
	GetPostByIDFunc     func(ctx context.Context, id int64) (*entity.Post, error)
	DeletePostFunc      func(ctx context.Context, postID, authorID int64, role string) error
	UpdatePostFunc      func(ctx context.Context, postID, authorID int64, role, title, content string) (*entity.Post, error)
	GetPostsByUserIDFunc func(ctx context.Context, userID int64) ([]*entity.Post, error)
}

func (m *MockPostRepository) CreatePost(ctx context.Context, post *entity.Post) (int64, error) {
	if m.CreatePostFunc != nil {
		return m.CreatePostFunc(ctx, post)
	}
	return 0, nil
}

func (m *MockPostRepository) GetPosts(ctx context.Context) ([]*entity.Post, error) {
	if m.GetPostsFunc != nil {
		return m.GetPostsFunc(ctx)
	}
	return nil, nil
}

func (m *MockPostRepository) GetPostByID(ctx context.Context, id int64) (*entity.Post, error) {
	if m.GetPostByIDFunc != nil {
		return m.GetPostByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockPostRepository) DeletePost(ctx context.Context, postID, authorID int64, role string) error {
	if m.DeletePostFunc != nil {
		return m.DeletePostFunc(ctx, postID, authorID, role)
	}
	return nil
}

func (m *MockPostRepository) UpdatePost(ctx context.Context, postID, authorID int64, role, title, content string) (*entity.Post, error) {
	if m.UpdatePostFunc != nil {
		return m.UpdatePostFunc(ctx, postID, authorID, role, title, content)
	}
	return nil, nil
}

func (m *MockPostRepository) GetPostsByUserID(ctx context.Context, userID int64) ([]*entity.Post, error) {
	if m.GetPostsByUserIDFunc != nil {
		return m.GetPostsByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

// Mock implementation of AuthServiceClient interface
type MockAuthServiceClient struct {
	ValidateTokenFunc    func(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error)
	GetUserProfileFunc   func(ctx context.Context, in *pb.GetUserProfileRequest, opts ...grpc.CallOption) (*pb.GetUserProfileResponse, error)
	LoginFunc           func(ctx context.Context, in *pb.LoginRequest, opts ...grpc.CallOption) (*pb.TokenResponse, error)
	RegisterFunc        func(ctx context.Context, req *pb.RegisterRequest, opts ...grpc.CallOption) (*pb.UserResponse, error)
	LogoutFunc          func(ctx context.Context, in *pb.LogoutRequest, opts ...grpc.CallOption) (*pb.SuccessResponse, error)
	RefreshTokenFunc    func(ctx context.Context, in *pb.RefreshTokenRequest, opts ...grpc.CallOption) (*pb.TokenResponse, error)
	SignInFunc          func(ctx context.Context, in *pb.SignInRequest, opts ...grpc.CallOption) (*pb.SignInResponse, error)
	SignUpFunc          func(ctx context.Context, in *pb.SignUpRequest, opts ...grpc.CallOption) (*pb.SignUpResponse, error)
	ValidateSessionFunc func(ctx context.Context, in *pb.ValidateSessionRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error)
}

func (m *MockAuthServiceClient) ValidateToken(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(ctx, in, opts...)
	}
	return nil, nil
}

func (m *MockAuthServiceClient) GetUserProfile(ctx context.Context, in *pb.GetUserProfileRequest, opts ...grpc.CallOption) (*pb.GetUserProfileResponse, error) {
	if m.GetUserProfileFunc != nil {
		return m.GetUserProfileFunc(ctx, in, opts...)
	}
	return nil, nil
}

func (m *MockAuthServiceClient) Login(ctx context.Context, in *pb.LoginRequest, opts ...grpc.CallOption) (*pb.TokenResponse, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(ctx, in, opts...)
	}
	return nil, nil
}

func (m *MockAuthServiceClient) Register(ctx context.Context, req *pb.RegisterRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, req, opts...)
	}
	return &pb.UserResponse{}, nil
}

func (m *MockAuthServiceClient) Logout(ctx context.Context, in *pb.LogoutRequest, opts ...grpc.CallOption) (*pb.SuccessResponse, error) {
	if m.LogoutFunc != nil {
		return m.LogoutFunc(ctx, in, opts...)
	}
	return nil, nil
}

func (m *MockAuthServiceClient) RefreshToken(ctx context.Context, in *pb.RefreshTokenRequest, opts ...grpc.CallOption) (*pb.TokenResponse, error) {
	if m.RefreshTokenFunc != nil {
		return m.RefreshTokenFunc(ctx, in, opts...)
	}
	return nil, nil
}

func (m *MockAuthServiceClient) SignIn(ctx context.Context, in *pb.SignInRequest, opts ...grpc.CallOption) (*pb.SignInResponse, error) {
	if m.SignInFunc != nil {
		return m.SignInFunc(ctx, in, opts...)
	}
	return &pb.SignInResponse{}, nil
}

func (m *MockAuthServiceClient) SignUp(ctx context.Context, in *pb.SignUpRequest, opts ...grpc.CallOption) (*pb.SignUpResponse, error) {
	if m.SignUpFunc != nil {
		return m.SignUpFunc(ctx, in, opts...)
	}
	return &pb.SignUpResponse{}, nil
}

func (m *MockAuthServiceClient) ValidateSession(ctx context.Context, in *pb.ValidateSessionRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
	if m.ValidateSessionFunc != nil {
		return m.ValidateSessionFunc(ctx, in, opts...)
	}
	return &pb.ValidateSessionResponse{}, nil
}