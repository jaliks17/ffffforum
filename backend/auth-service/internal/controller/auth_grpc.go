package controller

import (
	"context"

	pb "github.com/jaliks17/ffffforum/backend/proto"

	"github.com/jaliks17/ffffforum/backend/auth-service/internal/entity"
	"github.com/jaliks17/ffffforum/backend/auth-service/internal/usecase"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthGRPCController struct {
	pb.UnimplementedAuthServiceServer
	authUC usecase.IAuthUseCase
}

func NewAuthGRPCController(authUC usecase.IAuthUseCase) *AuthGRPCController {
	return &AuthGRPCController{authUC: authUC}
}

func (c *AuthGRPCController) SignUp(
	ctx context.Context,
	req *pb.SignUpRequest,
) (*pb.SignUpResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	user := entity.UserRegister{
		Username: req.Username,
		Password: req.Password,
	}

	createdUser, err := c.authUC.Register(ctx, user)
	if err != nil {
		if err == usecase.ErrUserExists {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Errorf(codes.Internal, "registration failed: %v", err)
	}

	return &pb.SignUpResponse{
		UserId: createdUser.ID,
	}, nil
}

func (c *AuthGRPCController) SignIn(
	ctx context.Context,
	req *pb.SignInRequest,
) (*pb.SignInResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	loginReq := entity.UserLogin{
		Username: req.Username,
		Password: req.Password,
	}

	session, err := c.authUC.Login(ctx, loginReq)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
	}

	return &pb.SignInResponse{
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
	}, nil
}

func (c *AuthGRPCController) GetUserProfile(
	ctx context.Context,
	req *pb.GetUserProfileRequest,
) (*pb.GetUserProfileResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	user, err := c.authUC.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &pb.GetUserProfileResponse{
		User: convertUserToProto(user),
	}, nil
}

func (c *AuthGRPCController) ValidateSession(
	ctx context.Context,
	req *pb.ValidateSessionRequest,
) (*pb.ValidateSessionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	token, err := c.authUC.ValidateToken(req.Token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid token claims")
	}

	return &pb.ValidateSessionResponse{
		Valid:    true,
		UserId:   int64(claims["user_id"].(float64)),
		UserRole: claims["role"].(string),
	}, nil
}

func (c *AuthGRPCController) Login(ctx context.Context, req *pb.LoginRequest) (*pb.TokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	loginReq := entity.UserLogin{
		Username: req.Username,
		Password: req.Password,
	}

	session, err := c.authUC.Login(ctx, loginReq)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
	}

	return &pb.TokenResponse{
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
	}, nil
}

func (c *AuthGRPCController) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.UserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	user := entity.UserRegister{
		Username: req.Username,
		Password: req.Password,
	}

	createdUser, err := c.authUC.Register(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "registration failed: %v", err)
	}

	return &pb.UserResponse{
		Id:       uint64(createdUser.ID),
		Username: createdUser.Username,
		Role:     string(createdUser.Role),
	}, nil
}

func (c *AuthGRPCController) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateSessionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	token, err := c.authUC.ValidateToken(req.Token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid token claims")
	}

	return &pb.ValidateSessionResponse{
		Valid:    true,
		UserId:   int64(claims["user_id"].(float64)),
		UserRole: claims["role"].(string),
	}, nil
}

func (c *AuthGRPCController) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.TokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	session, err := c.authUC.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "refresh token failed: %v", err)
	}

	return &pb.TokenResponse{
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
	}, nil
}

func (c *AuthGRPCController) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.SuccessResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	err := c.authUC.Logout(ctx, req.Token)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "logout failed: %v", err)
	}

	return &pb.SuccessResponse{
		Message: "Successfully logged out",
	}, nil
}

func convertUserToProto(user *entity.User) *pb.User {
	if user == nil {
		return nil
	}

	return &pb.User{
		Id:        user.ID,
		Username:  user.Username,
		Role:      string(user.Role),
		CreatedAt: timestamppb.New(user.CreatedAt),
	}
}