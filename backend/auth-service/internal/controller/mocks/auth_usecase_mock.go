package mocks

import (
	"context"

	"github.com/jaliks17/ffffforum/backend/auth-service/internal/entity"

	"github.com/stretchr/testify/mock"
)

type AuthUseCase struct {
	mock.Mock
}

func (m *AuthUseCase) Register(ctx context.Context, input entity.UserRegister) (*entity.User, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *AuthUseCase) Login(ctx context.Context, input entity.UserLogin) (*entity.TokenResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TokenResponse), args.Error(1)
}
