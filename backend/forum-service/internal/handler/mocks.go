package handler

import (
	"context"

	pb "backend/proto"

	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) ValidateToken(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.ValidateSessionResponse), args.Error(1)
}

func (m *MockAuthClient) GetUserProfile(ctx context.Context, in *pb.GetUserProfileRequest, opts ...grpc.CallOption) (*pb.GetUserProfileResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.GetUserProfileResponse), args.Error(1)
}

func (m *MockAuthClient) Login(ctx context.Context, in *pb.LoginRequest, opts ...grpc.CallOption) (*pb.TokenResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.TokenResponse), args.Error(1)
}

func (m *MockAuthClient) Register(ctx context.Context, in *pb.RegisterRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.UserResponse), args.Error(1)
}

func (m *MockAuthClient) Logout(ctx context.Context, in *pb.LogoutRequest, opts ...grpc.CallOption) (*pb.SuccessResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.SuccessResponse), args.Error(1)
}

func (m *MockAuthClient) RefreshToken(ctx context.Context, in *pb.RefreshTokenRequest, opts ...grpc.CallOption) (*pb.TokenResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.TokenResponse), args.Error(1)
}

func (m *MockAuthClient) SignIn(ctx context.Context, in *pb.SignInRequest, opts ...grpc.CallOption) (*pb.SignInResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.SignInResponse), args.Error(1)
}

func (m *MockAuthClient) SignUp(ctx context.Context, in *pb.SignUpRequest, opts ...grpc.CallOption) (*pb.SignUpResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.SignUpResponse), args.Error(1)
}

func (m *MockAuthClient) ValidateSession(ctx context.Context, in *pb.ValidateSessionRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.ValidateSessionResponse), args.Error(1)
} 