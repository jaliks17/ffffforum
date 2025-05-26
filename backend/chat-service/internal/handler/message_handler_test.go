package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/proto"
	"chat-service/internal/entity"
	myWeb "chat-service/pkg/websocket"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type MockMessageUseCase struct {
	mock.Mock
}

type MockAuthServiceClient struct {
	mock.Mock
}

func (m *MockMessageUseCase) SaveMessage(msg *entity.Message) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockMessageUseCase) GetMessages() ([]entity.Message, error) {
	args := m.Called()
	return args.Get(0).([]entity.Message), args.Error(1)
}

func (m *MockAuthServiceClient) GetUserProfile(ctx context.Context, in *proto.GetUserProfileRequest, opts ...grpc.CallOption) (*proto.GetUserProfileResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.GetUserProfileResponse), args.Error(1)
}

func (m *MockAuthServiceClient) Login(ctx context.Context, in *proto.LoginRequest, opts ...grpc.CallOption) (*proto.TokenResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.TokenResponse), args.Error(1)
}

func (m *MockAuthServiceClient) Register(ctx context.Context, in *proto.RegisterRequest, opts ...grpc.CallOption) (*proto.UserResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.UserResponse), args.Error(1)
}

func (m *MockAuthServiceClient) ValidateToken(ctx context.Context, in *proto.ValidateTokenRequest, opts ...grpc.CallOption) (*proto.ValidateSessionResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.ValidateSessionResponse), args.Error(1)
}

func (m *MockAuthServiceClient) RefreshToken(ctx context.Context, in *proto.RefreshTokenRequest, opts ...grpc.CallOption) (*proto.TokenResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.TokenResponse), args.Error(1)
}

func (m *MockAuthServiceClient) Logout(ctx context.Context, in *proto.LogoutRequest, opts ...grpc.CallOption) (*proto.SuccessResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.SuccessResponse), args.Error(1)
}

func (m *MockAuthServiceClient) SignIn(ctx context.Context, in *proto.SignInRequest, opts ...grpc.CallOption) (*proto.SignInResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.SignInResponse), args.Error(1)
}

func (m *MockAuthServiceClient) SignUp(ctx context.Context, in *proto.SignUpRequest, opts ...grpc.CallOption) (*proto.SignUpResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.SignUpResponse), args.Error(1)
}

func (m *MockAuthServiceClient) ValidateSession(ctx context.Context, in *proto.ValidateSessionRequest, opts ...grpc.CallOption) (*proto.ValidateSessionResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.ValidateSessionResponse), args.Error(1)
}

func TestMessageHandler_GetMessages(t *testing.T) {
	uc := new(MockMessageUseCase)
	authClient := new(MockAuthServiceClient)

	uc.On("GetMessages").Return([]entity.Message{
		{ID: 1, UserID: 1, Username: "testuser", Message: "Hello, World!"},
	}, nil)

	handler := NewMessageHandler(uc, authClient)

	router := gin.Default()
	router.GET("/messages", handler.GetMessages)

	req, _ := http.NewRequest("GET", "/messages", nil)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []entity.Message
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 1, len(resp))
	assert.Equal(t, "Hello, World!", resp[0].Message)
	assert.Equal(t, "testuser", resp[0].Username)
	uc.AssertExpectations(t)
}

func TestMessageHandler_HandleConnections(t *testing.T) {
	uc := new(MockMessageUseCase)
	authClient := new(MockAuthServiceClient)

	uc.On("SaveMessage", mock.Anything).Return(nil)

	handler := NewMessageHandler(uc, authClient)

	router := gin.Default()
	router.GET("/ws", handler.HandleConnections)

	server := httptest.NewServer(router)
	defer server.Close()

	url := "ws" + server.URL[4:] + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	defer ws.Close()

	msg := entity.Message{UserID: 1, Username: "testuser", Message: "Hello, World!"}
	err = ws.WriteJSON(msg)
	assert.NoError(t, err)

	uc.AssertExpectations(t)
}

func TestMessageHandler_HandleMessages(t *testing.T) {
	// Создаем мок usecase
	uc := new(MockMessageUseCase)
	authClient := new(MockAuthServiceClient)

	// Создаем MessageHandler
	handler := NewMessageHandler(uc, authClient)

	// Создаем канал для тестирования
	broadcast := make(chan entity.Message)
	myWeb.Broadcast = broadcast

	// Запускаем горутину для обработки сообщений
	go handler.HandleMessages()

	// Отправляем сообщение в канал
	msg := entity.Message{UserID: 1, Username: "testuser", Message: "Hello, World!"}
	broadcast <- msg

	// Проверяем, что сообщение было отправлено всем клиентам
	// (здесь можно добавить проверку для клиентов, если они есть)
}

func TestMessageHandler_GetMessages_Error(t *testing.T) {
	uc := new(MockMessageUseCase)
	authClient := new(MockAuthServiceClient)
	uc.On("GetMessages").Return(nil, errors.New("database error"))

	handler := NewMessageHandler(uc, authClient)
	router := gin.Default()
	router.GET("/messages", handler.GetMessages)

	req, _ := http.NewRequest("GET", "/messages", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	uc.AssertExpectations(t)
}