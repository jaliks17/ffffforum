package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/entity"
	"auth-service/internal/repository"
	"auth-service/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("неверные учетные данные")
	ErrUserNotFound      = errors.New("пользователь не найден")
	ErrUserExists        = errors.New("пользователь уже существует")
	ErrInvalidUsername   = errors.New("неверный формат имени пользователя")
	ErrInvalidPassword   = errors.New("пароль должен содержать минимум 6 символов")
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,}$`)

type IAuthUseCase interface {
	Register(ctx context.Context, input entity.UserRegister) (*entity.User, error)
	Login(ctx context.Context, input entity.UserLogin) (*entity.TokenResponse, error)
	GetUserByID(ctx context.Context, id int64) (*entity.User, error)
	ValidateToken(token string) (*jwt.Token, error)
	RefreshToken(ctx context.Context, refreshToken string) (*entity.TokenResponse, error)
	Logout(ctx context.Context, token string) error
}

type AuthUseCase struct {
	userRepo    repository.IUserRepository
	sessionRepo repository.ISessionRepository
	config      *config.AuthConfig
	logger      *logger.Logger
}

func NewAuthUseCase(
	userRepo repository.IUserRepository,
	sessionRepo repository.ISessionRepository,
	config *config.AuthConfig,
	logger *logger.Logger,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		config:      config,
		logger:      logger,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, input entity.UserRegister) (*entity.User, error) {
	// Проверяем формат username
	if !usernameRegex.MatchString(input.Username) {
		return nil, ErrInvalidUsername
	}

	// Проверяем длину пароля
	if len(input.Password) < 6 {
		return nil, ErrInvalidPassword
	}

	// Проверяем, существует ли пользователь
	existingUser, err := uc.userRepo.GetByUsername(ctx, input.Username)
	if err == nil && existingUser != nil {
		return nil, ErrUserExists
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Логируем хеш для отладки
	uc.logger.Debug("Register: password hash generated",
		zap.String("username", input.Username),
		zap.String("password", input.Password),
		zap.String("hashed_password_bytes", fmt.Sprintf("%v", hashedPassword)), // Log byte slice
		zap.String("hashed_string", string(hashedPassword)),                 // Log string conversion of bytes
		zap.String("hashed_hex_representation", fmt.Sprintf("%x", hashedPassword))) // Log hex representation for comparison

	// Создаем нового пользователя
	user := &entity.User{
		Username: input.Username,
		Password: string(hashedPassword), // <-- Save as raw bcrypt string
		Role:     "user",
	}

	id, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	user.ID = id

	return user, nil
}

func (uc *AuthUseCase) Login(ctx context.Context, input entity.UserLogin) (*entity.TokenResponse, error) {
	// Получаем пользователя по username
	user, err := uc.userRepo.GetByUsername(ctx, input.Username)
	if err != nil {
		uc.logger.Error("Login failed: error getting user by username", zap.Error(err), zap.String("username", input.Username))
		return nil, errors.New("internal server error") // Возвращаем общую ошибку при ошибке БД
	}

	// Логируем состояние пользователя после получения из репозитория
	uc.logger.Debug("Login: user state after repo call",
		zap.Any("user_object", user),
		zap.Bool("user_is_nil_check", user == nil))

	// Если пользователь не найден
	if user == nil {
		uc.logger.Warn("Login failed: user not found", zap.String("username", input.Username))
		return nil, ErrInvalidCredentials
	}

	// Используем хеш пароля из базы данных напрямую (как raw bcrypt string)
	hashedPasswordFromDB := []byte(user.Password)

	// Логируем детали хеша из БД
	uc.logger.Debug("Login: hash details from DB",
		zap.String("hash_from_db_string", user.Password), // Log original string from DB
		zap.Int("hash_string_len", len(user.Password)),
		zap.Int("hash_bytes_len", len(hashedPasswordFromDB)), // Log length of byte slice
		zap.String("hash_bytes", fmt.Sprintf("%v", hashedPasswordFromDB))) // Log byte slice

	// Сравниваем введенный пароль с хешем из базы данных
	if err := bcrypt.CompareHashAndPassword(hashedPasswordFromDB, []byte(input.Password)); err != nil {
		// Если пароль не совпадает
		uc.logger.Warn("Login failed: invalid password", zap.String("username", input.Username), zap.Error(err))
		return nil, ErrInvalidCredentials // Возвращаем ошибку неверных учетных данных
	}

	// Логирование после успешного сравнения паролей
	uc.logger.Debug("Login: password comparison successful, proceeding to token generation",
		zap.String("username", input.Username))

	// Логируем перед созданием токена
	uc.logger.Debug("Login: creating JWT token", zap.Int64("user_id", user.ID), zap.String("username", user.Username), zap.String("role", user.Role))

	// Создаем JWT токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(uc.config.Expiration).Unix(), // Use config expiration
	})

	// Логируем JWT секрет перед подписанием
	uc.logger.Debug("Login: JWT secret details before signing",
		zap.Int("secret_length", len(uc.config.Secret))) // Log length only, not value

	// Подписываем основной токен пользователя
	tokenString, err := token.SignedString([]byte(uc.config.Secret))
	if err != nil {
		uc.logger.Error("Login failed: failed to sign JWT token", zap.Error(err), zap.String("username", input.Username))
		return nil, errors.New("internal server error") // Возвращаем общую ошибку при ошибке подписи
	}

	// Логируем перед успешным возвратом
	uc.logger.Debug("Login: token signed successfully, returning response",
		zap.String("username", input.Username))

	// TODO: Generate Refresh Token and save session

	return &entity.TokenResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   int64(uc.config.Expiration.Seconds()), // Use config expiration in seconds
		RefreshToken: "", // Placeholder, implement refresh token logic
	}, nil
}

func (uc *AuthUseCase) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неверный метод подписи токена")
		}
		return []byte(uc.config.Secret), nil // Преобразуем секрет в байтовый срез
	})
}

func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (*entity.TokenResponse, error) {
	// TODO: Реализовать логику обновления токена
	return nil, errors.New("not implemented")
}

func (uc *AuthUseCase) Logout(ctx context.Context, token string) error {
	// TODO: Реализовать логику выхода
	return errors.New("not implemented")
}

func (uc *AuthUseCase) GetUserByID(ctx context.Context, id int64) (*entity.User, error) {
	// Fetch user from repository by ID
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		// Handle repository errors other than sql.ErrNoRows
		uc.logger.Error("GetUserByID failed: repository error", zap.Error(err), zap.Int64("user_id", id))
		return nil, errors.New("internal server error") // Return a generic error for unexpected repo errors
	}

	// Check if user was not found (repository returned nil for sql.ErrNoRows)
	if user == nil {
		return nil, ErrUserNotFound // Return the use case specific error
	}

	return user, nil
}
