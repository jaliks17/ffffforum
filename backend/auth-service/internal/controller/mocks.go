package controller

import (
	"context"

	"auth-service/internal/entity"
	"auth-service/internal/usecase"

	"github.com/golang-jwt/jwt/v5"
)

// AuthServiceMock представляет мок-реализацию сервиса аутентификации
// для использования в unit-тестах
type AuthServiceMock struct{}

// Проверяем, что мок реализует интерфейс IAuthUseCase
var _ usecase.IAuthUseCase = (*AuthServiceMock)(nil)

// Реализация методов интерфейса AuthUseCase
func (m *AuthServiceMock) Register(ctx context.Context, input entity.UserRegister) (*entity.User, error) {
	return nil, nil
}

func (m *AuthServiceMock) Login(ctx context.Context, input entity.UserLogin) (*entity.TokenResponse, error) {
	return nil, nil
}

func (m *AuthServiceMock) GetUserByID(ctx context.Context, id int64) (*entity.User, error) {
	return nil, nil
}

func (m *AuthServiceMock) ValidateToken(token string) (*jwt.Token, error) {
	return nil, nil
}

func (m *AuthServiceMock) RefreshToken(ctx context.Context, refreshToken string) (*entity.TokenResponse, error) {
	return nil, nil
}

func (m *AuthServiceMock) Logout(ctx context.Context, token string) error {
	return nil
}