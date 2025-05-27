package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"auth-service/internal/entity"
	"auth-service/internal/usecase"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// AuthUseCase defines the interface for authentication use cases
type AuthUseCase interface {
	Register(ctx context.Context, input entity.UserRegister) (*entity.User, error)
	Login(ctx context.Context, input entity.UserLogin) (*entity.TokenResponse, error)
	GetUserByID(ctx context.Context, id int64) (*entity.User, error)
	ValidateToken(token string) (*jwt.Token, error)
	RefreshToken(ctx context.Context, refreshToken string) (*entity.TokenResponse, error)
	Logout(ctx context.Context, token string) error
}

type MockAuthUseCase struct {
	mock.Mock
}

func (m *MockAuthUseCase) Register(ctx context.Context, input entity.UserRegister) (*entity.User, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockAuthUseCase) Login(ctx context.Context, input entity.UserLogin) (*entity.TokenResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TokenResponse), args.Error(1)
}

func (m *MockAuthUseCase) GetUserByID(ctx context.Context, id int64) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockAuthUseCase) ValidateToken(token string) (*jwt.Token, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *MockAuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (*entity.TokenResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TokenResponse), args.Error(1)
}

func (m *MockAuthUseCase) Logout(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func setupTestRouter(mockUC *MockAuthUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	controller := NewAuthHTTPController(mockUC)

	router.POST("/api/v1/auth/signup", controller.SignUp)
	router.POST("/api/v1/auth/signin", controller.SignIn)
	router.GET("/api/v1/auth/users/:id", controller.GetUserProfile)
	router.GET("/api/v1/auth/validate", controller.ValidateToken)

	return router
}

func TestSignUp(t *testing.T) {
	mockUC := new(MockAuthUseCase)
	router := setupTestRouter(mockUC)

	tests := []struct {
		name           string
		payload        SignUpRequest
		mockSetup      func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful registration",
			payload: SignUpRequest{
				Username: "testuser",
				Password: "password123",
				Role:     "user",
			},
			mockSetup: func() {
				mockUC.On("Register", mock.Anything, entity.UserRegister{
					Username: "testuser",
					Password: "password123",
					Role:     "user",
				}).Return(&entity.User{ID: 1}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"id": float64(1),
			},
		},
		{
			name: "invalid request body",
			payload: SignUpRequest{
				Username: "te", // too short
				Password: "123",
				Role:     "invalid",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": mock.Anything,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedStatus == http.StatusCreated {
				assert.Equal(t, tt.expectedBody["id"], response["id"])
			} else {
				assert.Contains(t, response, "error")
			}
		})
	}
}

func TestSignIn(t *testing.T) {
	mockUC := new(MockAuthUseCase)
	router := setupTestRouter(mockUC)

	tests := []struct {
		name           string
		payload        SignInRequest
		mockSetup      func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful login",
			payload: SignInRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockSetup: func() {
				mockUC.On("Login", mock.Anything, entity.UserLogin{
					Username: "testuser",
					Password: "password123",
				}).Return(&entity.TokenResponse{
					AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6InRlc3R1c2VyIiwicm9sZSI6InVzZXIifQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
					RefreshToken: "test-refresh-token",
					TokenType:    "Bearer",
					ExpiresIn:    86400,
				}, nil)
				mockUC.On("GetUserByID", mock.Anything, int64(1)).Return(&entity.User{
					ID:       1,
					Username: "testuser",
					Role:     "user",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"access_token":  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6InRlc3R1c2VyIiwicm9sZSI6InVzZXIifQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
				"refresh_token": "test-refresh-token",
			},
		},
		{
			name: "invalid credentials",
			payload: SignInRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			mockSetup: func() {
				mockUC.On("Login", mock.Anything, entity.UserLogin{
					Username: "testuser",
					Password: "wrongpassword",
				}).Return(nil, usecase.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "invalid credentials",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signin", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.expectedBody["access_token"], response["access_token"])
				assert.Equal(t, tt.expectedBody["refresh_token"], response["refresh_token"])
			} else {
				assert.Equal(t, tt.expectedBody["error"], response["error"])
			}
		})
	}
}

func TestGetUserProfile(t *testing.T) {
	mockUC := new(MockAuthUseCase)
	router := setupTestRouter(mockUC)

	tests := []struct {
		name           string
		userID         string
		mockSetup      func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "user found",
			userID: "1",
			mockSetup: func() {
				mockUC.On("GetUserByID", mock.Anything, int64(1)).Return(&entity.User{
					ID:       1,
					Username: "testuser",
					Role:     "user",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":        float64(1),
				"username":  "testuser",
				"role":      "user",
				"password":  "",
			},
		},
		{
			name:   "user not found",
			userID: "999",
			mockSetup: func() {
				mockUC.On("GetUserByID", mock.Anything, int64(999)).Return(nil, usecase.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "user not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/users/"+tt.userID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.expectedBody["id"], response["id"])
				assert.Equal(t, tt.expectedBody["username"], response["username"])
				assert.Equal(t, tt.expectedBody["role"], response["role"])
			} else {
				assert.Equal(t, tt.expectedBody["error"], response["error"])
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	mockUC := new(MockAuthUseCase)
	router := setupTestRouter(mockUC)

	tests := []struct {
		name           string
		token          string
		mockSetup      func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:  "valid token",
			token: "valid-token",
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
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"user_id":  float64(1),
				"role":     "user",
			},
		},
		{
			name:           "missing token",
			token:          "",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "missing auth token",
			},
		},
		{
			name:  "invalid token",
			token: "invalid-token",
			mockSetup: func() {
				mockUC.On("ValidateToken", "invalid-token").Return(nil, errors.New("invalid token"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "invalid token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/validate", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.expectedBody["user_id"], response["user_id"])
				assert.Equal(t, tt.expectedBody["role"], response["role"])
			} else {
				assert.Equal(t, tt.expectedBody["error"], response["error"])
			}
		})
	}
}
