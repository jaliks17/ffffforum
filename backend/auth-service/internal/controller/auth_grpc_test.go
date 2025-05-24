package controller

import (
	"context"
	"testing"

	"auth-service/internal/entity"
	pb "backend/proto"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthGRPCController_SignUp(t *testing.T) {
	mockUC := new(MockAuthUseCase)
	ctrl := NewAuthGRPCController(mockUC)

	tests := []struct {
		name           string
		req            *pb.SignUpRequest
		mockSetup      func()
		expectedError  bool
		expectedUserID int64
	}{
		{
			name: "successful registration",
			req: &pb.SignUpRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockSetup: func() {
				mockUC.On("Register", mock.Anything, entity.UserRegister{
					Username: "testuser",
					Password: "password123",
				}).Return(&entity.User{ID: 1}, nil)
			},
			expectedError:  false,
			expectedUserID: 1,
		},
		{
			name:           "empty request",
			req:            nil,
			mockSetup:      func() {},
			expectedError:  true,
			expectedUserID: 0,
		},
		{
			name: "registration failed",
			req: &pb.SignUpRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockSetup: func() {
				mockUC.On("Register", mock.Anything, entity.UserRegister{
					Username: "testuser",
					Password: "password123",
				}).Return(nil, assert.AnError)
			},
			expectedError:  true,
			expectedUserID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := ctrl.SignUp(context.Background(), tt.req)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedUserID, resp.UserId)
			}
		})
	}
}

func TestAuthGRPCController_SignIn(t *testing.T) {
	mockUC := new(MockAuthUseCase)
	ctrl := NewAuthGRPCController(mockUC)

	tests := []struct {
		name           string
		req            *pb.SignInRequest
		mockSetup      func()
		expectedError  bool
		expectedTokens *entity.TokenResponse
	}{
		{
			name: "successful login",
			req: &pb.SignInRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockSetup: func() {
				mockUC.On("Login", mock.Anything, entity.UserLogin{
					Username: "testuser",
					Password: "password123",
				}).Return(&entity.TokenResponse{
					AccessToken:  "test-access-token",
					RefreshToken: "test-refresh-token",
				}, nil)
			},
			expectedError: false,
			expectedTokens: &entity.TokenResponse{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
			},
		},
		{
			name:           "empty request",
			req:            nil,
			mockSetup:      func() {},
			expectedError:  true,
			expectedTokens: nil,
		},
		{
			name: "login failed",
			req: &pb.SignInRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			mockSetup: func() {
				mockUC.On("Login", mock.Anything, entity.UserLogin{
					Username: "testuser",
					Password: "wrongpassword",
				}).Return(nil, assert.AnError)
			},
			expectedError:  true,
			expectedTokens: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := ctrl.SignIn(context.Background(), tt.req)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedTokens.AccessToken, resp.AccessToken)
				assert.Equal(t, tt.expectedTokens.RefreshToken, resp.RefreshToken)
			}
		})
	}
}

func TestAuthGRPCController_GetUserProfile(t *testing.T) {
	mockUC := new(MockAuthUseCase)
	ctrl := NewAuthGRPCController(mockUC)

	tests := []struct {
		name          string
		req           *pb.GetUserProfileRequest
		mockSetup     func()
		expectedError bool
		expectedUser  *entity.User
	}{
		{
			name: "user found",
			req: &pb.GetUserProfileRequest{
				UserId: 1,
			},
			mockSetup: func() {
				mockUC.On("GetUserByID", mock.Anything, int64(1)).Return(&entity.User{
					ID:       1,
					Username: "testuser",
					Role:     "user",
				}, nil)
			},
			expectedError: false,
			expectedUser: &entity.User{
				ID:       1,
				Username: "testuser",
				Role:     "user",
			},
		},
		{
			name:           "empty request",
			req:            nil,
			mockSetup:      func() {},
			expectedError:  true,
			expectedUser:   nil,
		},
		{
			name: "user not found",
			req: &pb.GetUserProfileRequest{
				UserId: 999,
			},
			mockSetup: func() {
				mockUC.On("GetUserByID", mock.Anything, int64(999)).Return(nil, assert.AnError)
			},
			expectedError: true,
			expectedUser:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := ctrl.GetUserProfile(context.Background(), tt.req)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedUser.ID, resp.User.Id)
				assert.Equal(t, tt.expectedUser.Username, resp.User.Username)
				assert.Equal(t, tt.expectedUser.Role, resp.User.Role)
			}
		})
	}
}

func TestAuthGRPCController_ValidateSession(t *testing.T) {
	mockUC := new(MockAuthUseCase)
	ctrl := NewAuthGRPCController(mockUC)

	tests := []struct {
		name          string
		req           *pb.ValidateSessionRequest
		mockSetup     func()
		expectedError bool
		expectedValid bool
	}{
		{
			name: "valid token",
			req: &pb.ValidateSessionRequest{
				Token: "valid-token",
			},
			mockSetup: func() {
				token := &jwt.Token{
					Claims: jwt.MapClaims{
						"user_id":  float64(1),
						"username": "testuser",
						"role":     "user",
					},
				}
				mockUC.On("ValidateToken", "valid-token").Return(token, nil)
			},
			expectedError: false,
			expectedValid: true,
		},
		{
			name:           "empty request",
			req:            nil,
			mockSetup:      func() {},
			expectedError:  true,
			expectedValid: false,
		},
		{
			name: "invalid token",
			req: &pb.ValidateSessionRequest{
				Token: "invalid-token",
			},
			mockSetup: func() {
				mockUC.On("ValidateToken", "invalid-token").Return(nil, assert.AnError)
			},
			expectedError: true,
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := ctrl.ValidateSession(context.Background(), tt.req)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedValid, resp.Valid)
			}
		})
	}
}
