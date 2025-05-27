package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/entity"
	"auth-service/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) (int64, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *entity.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByToken(ctx context.Context, token string) (*entity.Session, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Session), args.Error(1)
}

func (m *MockSessionRepository) Delete(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestRegister(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger, _ := logger.NewLogger("info")
	config := &config.AuthConfig{
		Secret:     "test-secret",
		Expiration: time.Hour * 24,
	}

	uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, config, logger)

	tests := []struct {
		name          string
		input         entity.UserRegister
		mockSetup     func()
		expectedError error
	}{
		{
			name: "successful registration",
			input: entity.UserRegister{
				Username: "testuser",
				Password: "password123",
			},
			mockSetup: func() {
				mockUserRepo.On("GetByUsername", mock.Anything, "testuser").Return(nil, nil)
				mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(1), nil)
			},
			expectedError: nil,
		},
		{
			name: "invalid username format",
			input: entity.UserRegister{
				Username: "in",
				Password: "password123",
			},
			mockSetup:     func() {},
			expectedError: ErrInvalidUsername,
		},
		{
			name: "password too short",
			input: entity.UserRegister{
				Username: "testuser",
				Password: "12345",
			},
			mockSetup:     func() {},
			expectedError: ErrInvalidPassword,
		},
		{
			name: "user already exists",
			input: entity.UserRegister{
				Username: "existinguser",
				Password: "password123",
			},
			mockSetup: func() {
				mockUserRepo.On("GetByUsername", mock.Anything, "existinguser").Return(&entity.User{}, nil)
			},
			expectedError: ErrUserExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			user, err := uc.Register(context.Background(), tt.input)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.input.Username, user.Username)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger, _ := logger.NewLogger("info")
	config := &config.AuthConfig{
		Secret:     "test-secret",
		Expiration: time.Hour * 24,
	}

	uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, config, logger)

	tests := []struct {
		name          string
		input         entity.UserLogin
		mockSetup     func()
		expectedError error
	}{
		{
			name: "successful login",
			input: entity.UserLogin{
				Username: "testuser",
				Password: "password123",
			},
			mockSetup: func() {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				mockUserRepo.On("GetByUsername", mock.Anything, "testuser").Return(&entity.User{
					ID:       1,
					Username: "testuser",
					Password: string(hashedPassword),
					Role:     "user",
				}, nil)
			},
			expectedError: nil,
		},
		{
			name: "user not found",
			input: entity.UserLogin{
				Username: "nonexistentuser",
				Password: "password123",
			},
			mockSetup: func() {
				mockUserRepo.On("GetByUsername", mock.Anything, "nonexistentuser").Return(nil, nil)
			},
			expectedError: ErrInvalidCredentials,
		},
		{
			name: "invalid password",
			input: entity.UserLogin{
				Username: "testuser",
				Password: "wrongpassword",
			},
			mockSetup: func() {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				mockUserRepo.On("GetByUsername", mock.Anything, "testuser").Return(&entity.User{
					ID:       1,
					Username: "testuser",
					Password: string(hashedPassword),
					Role:     "user",
				}, nil)
			},
			expectedError: ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			token, err := uc.Login(context.Background(), tt.input)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, token)
				assert.NotEmpty(t, token.AccessToken)
				assert.Equal(t, "Bearer", token.TokenType)
				assert.Equal(t, int64(86400), token.ExpiresIn)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger, _ := logger.NewLogger("info")
	config := &config.AuthConfig{
		Secret:     "test-secret",
		Expiration: time.Hour * 24,
	}

	uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, config, logger)

	// Создаем валидный токен через Login
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	mockUserRepo.On("GetByUsername", mock.Anything, "testuser").Return(&entity.User{
		ID:       1,
		Username: "testuser",
		Password: string(hashedPassword),
		Role:     "user",
	}, nil)
	validToken, _ := uc.Login(context.Background(), entity.UserLogin{
		Username: "testuser",
		Password: "password123",
	})

	tests := []struct {
		name          string
		token         string
		expectedError bool
	}{
		{
			name:          "empty token",
			token:         "",
			expectedError: true,
		},
		{
			name:          "invalid token format",
			token:         "invalid.token.format",
			expectedError: true,
		},
		{
			name:          "valid token",
			token:         validToken.AccessToken,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := uc.ValidateToken(tt.token)
			if tt.expectedError {
				assert.Error(t, err)
				if token != nil {
					assert.False(t, token.Valid)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, token)
				assert.True(t, token.Valid)
			}
		})
	}
}

func TestRefreshToken(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger, _ := logger.NewLogger("info")
	config := &config.AuthConfig{
		Secret:     "test-secret",
		Expiration: time.Hour * 24,
	}

	uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, config, logger)

	tests := []struct {
		name          string
		refreshToken  string
		mockSetup     func()
		expectedError bool
	}{
		{
			name:         "not implemented",
			refreshToken: "valid-refresh-token",
			mockSetup:    func() {},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			token, err := uc.RefreshToken(context.Background(), tt.refreshToken)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, token)
			}
		})
	}
}

func TestLogout(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger, _ := logger.NewLogger("info")
	config := &config.AuthConfig{
		Secret:     "test-secret",
		Expiration: time.Hour * 24,
	}

	uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, config, logger)

	tests := []struct {
		name          string
		token         string
		mockSetup     func()
		expectedError bool
	}{
		{
			name:         "not implemented",
			token:        "valid-token",
			mockSetup:    func() {},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := uc.Logout(context.Background(), tt.token)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger, _ := logger.NewLogger("info")
	config := &config.AuthConfig{
		Secret:     "test-secret",
		Expiration: time.Hour * 24,
	}

	uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, config, logger)

	tests := []struct {
		name          string
		userID        int64
		mockSetup     func()
		expectedError error
	}{
		{
			name:   "user found",
			userID: 1,
			mockSetup: func() {
				mockUserRepo.On("GetByID", mock.Anything, int64(1)).Return(&entity.User{
					ID:       1,
					Username: "testuser",
					Role:     "user",
				}, nil)
			},
			expectedError: nil,
		},
		{
			name:   "user not found",
			userID: 999,
			mockSetup: func() {
				mockUserRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, nil)
			},
			expectedError: ErrUserNotFound,
		},
		{
			name:   "repository error",
			userID: 1,
			mockSetup: func() {
				mockUserRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, assert.AnError)
			},
			expectedError: errors.New("internal server error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo.ExpectedCalls = nil
			tt.mockSetup()
			user, err := uc.GetUserByID(context.Background(), tt.userID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.userID, user.ID)
			}
		})
	}
}
