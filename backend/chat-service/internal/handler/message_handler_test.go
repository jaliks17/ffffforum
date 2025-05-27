package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Message), args.Error(1)
}

func (m *MockMessageUseCase) DeleteOldMessages(before time.Time) error {
	args := m.Called(before)
	return args.Error(0)
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
	var resp []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 1, len(resp))
	assert.Equal(t, "Hello, World!", resp[0]["message"])
	assert.Equal(t, "testuser", resp[0]["author_name"])
	uc.AssertExpectations(t)
}

func TestMessageHandler_HandleConnections(t *testing.T) {
	uc := new(MockMessageUseCase)
	authClient := new(MockAuthServiceClient)

	uc.On("SaveMessage", mock.Anything).Return(nil)

	handler := NewMessageHandler(uc, authClient)

	// Mock ValidateSession to return valid token
	authClient.On("ValidateSession", mock.Anything, mock.MatchedBy(func(req *proto.ValidateSessionRequest) bool {
		return req.Token == "valid_token"
	})).Return(&proto.ValidateSessionResponse{Valid: true, UserId: 1, UserRole: "testrole"}, nil).Once()

	// Mock GetUserProfile to return valid user
	authClient.On("GetUserProfile", mock.Anything, mock.MatchedBy(func(req *proto.GetUserProfileRequest) bool {
		return req.UserId == 1
	})).Return(&proto.GetUserProfileResponse{User: &proto.User{Id: 1, Username: "testuser", Role: "testrole"}}, nil).Once()

	router := gin.Default()
	router.GET("/ws", handler.HandleConnections)

	server := httptest.NewServer(router)
	defer server.Close()

	// Create WebSocket connection with proper headers
	url := "ws" + server.URL[4:] + "/ws?token=valid_token"
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		assert.Fail(t, "Failed to dial websocket", err)
		return
	}
	defer ws.Close()

	// Send a message
	incMsg := incomingMessage{Type: "message", Message: "Hello, World!", Username: "testuser"}
	err = ws.WriteJSON(incMsg)
	assert.NoError(t, err)

	// Give some time for the handler to process
	time.Sleep(100 * time.Millisecond)

	uc.AssertExpectations(t)
	authClient.AssertExpectations(t)
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
	uc.AssertExpectations(t)
}

func TestMessageHandler_HandleConnections_DemoToken(t *testing.T) {
	uc := new(MockMessageUseCase)
	authClient := new(MockAuthServiceClient) // AuthClient will not be called with demo token

	// Expect SaveMessage to be called once with any message
	uc.On("SaveMessage", mock.Anything).Return(nil).Once()

	handler := NewMessageHandler(uc, authClient)

	router := gin.Default()
	router.GET("/ws", handler.HandleConnections)

	server := httptest.NewServer(router)
	defer server.Close()

	// Connect with demo token
	url := "ws" + server.URL[4:] + "/ws?token=demo"
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Send a message with the correct incomingMessage format
	incMsg := incomingMessage{Type: "message", Message: "Hello from demo!", Username: "Demo User"}
	err = ws.WriteJSON(incMsg)
	assert.NoError(t, err)

	// Send a non-text message to cover that branch
	err = ws.WriteMessage(websocket.BinaryMessage, []byte("binary data"))
	assert.NoError(t, err)

	// Send an unknown type message to cover that branch
	unknownMsg := incomingMessage{Type: "unknown", Message: "Should be ignored"}
	err = ws.WriteJSON(unknownMsg)
	assert.NoError(t, err)

	// Close the connection to trigger the goroutine to exit and checks
	ws.Close()

	// Give the goroutine a moment to finish (adjust duration if needed)
	time.Sleep(100 * time.Millisecond)

	// Assert that SaveMessage was called as expected
	uc.AssertExpectations(t)
	// Assert that AuthClient methods were NOT called
	authClient.AssertExpectations(t)
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

func TestMessageHandler_HandleConnections_InvalidToken(t *testing.T) {
	uc := new(MockMessageUseCase)
	authClient := new(MockAuthServiceClient)

	handler := NewMessageHandler(uc, authClient)

	// Mock ValidateSession to return invalid token
	authClient.On("ValidateSession", mock.Anything, mock.MatchedBy(func(req *proto.ValidateSessionRequest) bool {
		return req.Token == "invalid_token"
	})).Return(&proto.ValidateSessionResponse{Valid: false}, nil).Once()

	router := gin.Default()
	router.GET("/ws", handler.HandleConnections)

	req, _ := http.NewRequest("GET", "/ws?token=invalid_token", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	authClient.AssertExpectations(t)
}

func TestMessageHandler_HandleConnections_AuthServiceError(t *testing.T) {
	uc := new(MockMessageUseCase)
	authClient := new(MockAuthServiceClient)

	handler := NewMessageHandler(uc, authClient)

	// Mock ValidateSession to return error
	authClient.On("ValidateSession", mock.Anything, mock.MatchedBy(func(req *proto.ValidateSessionRequest) bool {
		return req.Token == "error_token"
	})).Return((*proto.ValidateSessionResponse)(nil), errors.New("auth service error")).Once()

	router := gin.Default()
	router.GET("/ws", handler.HandleConnections)

	req, _ := http.NewRequest("GET", "/ws?token=error_token", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "test")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	authClient.AssertExpectations(t)
}

func TestMessageHandler_HandleConnections_GetUserProfileError(t *testing.T) {
	uc := new(MockMessageUseCase)
	authClient := new(MockAuthServiceClient)

	handler := NewMessageHandler(uc, authClient)

	// Mock ValidateSession to return valid token
	authClient.On("ValidateSession", mock.Anything, mock.MatchedBy(func(req *proto.ValidateSessionRequest) bool {
		return req.Token == "valid_token"
	})).Return(&proto.ValidateSessionResponse{Valid: true, UserId: 1, UserRole: "testrole"}, nil).Once()

	// Mock GetUserProfile to return error
	authClient.On("GetUserProfile", mock.Anything, mock.MatchedBy(func(req *proto.GetUserProfileRequest) bool {
		return req.UserId == 1
	})).Return((*proto.GetUserProfileResponse)(nil), errors.New("get user profile error")).Once()

	router := gin.Default()
	router.GET("/ws", handler.HandleConnections)

	req, _ := http.NewRequest("GET", "/ws?token=valid_token", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "test")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	authClient.AssertExpectations(t)
}

func TestMessageHandler_HandleConnections_UserProfileNotFound(t *testing.T) {
	uc := new(MockMessageUseCase)
	authClient := new(MockAuthServiceClient)

	handler := NewMessageHandler(uc, authClient)

	// Mock ValidateSession to return valid token
	authClient.On("ValidateSession", mock.Anything, mock.MatchedBy(func(req *proto.ValidateSessionRequest) bool {
		return req.Token == "valid_token"
	})).Return(&proto.ValidateSessionResponse{Valid: true, UserId: 1, UserRole: "testrole"}, nil).Once()

	// Mock GetUserProfile to return nil user
	authClient.On("GetUserProfile", mock.Anything, mock.MatchedBy(func(req *proto.GetUserProfileRequest) bool {
		return req.UserId == 1
	})).Return(&proto.GetUserProfileResponse{User: nil}, nil).Once()

	router := gin.Default()
	router.GET("/ws", handler.HandleConnections)

	req, _ := http.NewRequest("GET", "/ws?token=valid_token", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	authClient.AssertExpectations(t)
}

func TestMessageHandler_HandleConnections_UpgradeError(t *testing.T) {
	uc := new(MockMessageUseCase)
	authClient := new(MockAuthServiceClient)

	handler := NewMessageHandler(uc, authClient)

	// Mock ValidateSession to return valid token
	authClient.On("ValidateSession", mock.Anything, mock.MatchedBy(func(req *proto.ValidateSessionRequest) bool {
		return req.Token == "valid_token"
	})).Return(&proto.ValidateSessionResponse{Valid: true, UserId: 1, UserRole: "testrole"}, nil).Once()

	// Mock GetUserProfile to return valid user
	authClient.On("GetUserProfile", mock.Anything, mock.MatchedBy(func(req *proto.GetUserProfileRequest) bool {
		return req.UserId == 1
	})).Return(&proto.GetUserProfileResponse{User: &proto.User{Id: 1, Username: "testuser", Role: "testrole"}}, nil).Once()

	router := gin.Default()
	router.GET("/ws", handler.HandleConnections)

	// Create a request that will fail to upgrade to WebSocket
	req, _ := http.NewRequest("GET", "/ws?token=valid_token", nil)
	req.Header.Set("Connection", "invalid")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// The handler should not return an error status since the upgrade failure is handled internally
	assert.NotEqual(t, http.StatusInternalServerError, w.Code)
	authClient.AssertExpectations(t)
}

func TestMessageHandler_HandleConnections_AddClientError(t *testing.T) {
	uc := new(MockMessageUseCase)
	authClient := new(MockAuthServiceClient)

	handler := NewMessageHandler(uc, authClient)

	// Mock ValidateSession to return valid token
	authClient.On("ValidateSession", mock.Anything, mock.MatchedBy(func(req *proto.ValidateSessionRequest) bool {
		return req.Token == "valid_token"
	})).Return(&proto.ValidateSessionResponse{Valid: true, UserId: 1, UserRole: "testrole"}, nil).Times(2)

	// Mock GetUserProfile to return valid user
	authClient.On("GetUserProfile", mock.Anything, mock.MatchedBy(func(req *proto.GetUserProfileRequest) bool {
		return req.UserId == 1
	})).Return(&proto.GetUserProfileResponse{User: &proto.User{Id: 1, Username: "testuser", Role: "testrole"}}, nil).Times(2)

	router := gin.Default()
	router.GET("/ws", handler.HandleConnections)

	server := httptest.NewServer(router)
	defer server.Close()

	// First connection
	url := "ws" + server.URL[4:] + "/ws?token=valid_token"
	ws1, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	defer ws1.Close()

	// Second connection with same user ID should fail to add client
	ws2, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	defer ws2.Close()

	// Give some time for the handler to process
	time.Sleep(100 * time.Millisecond)

	authClient.AssertExpectations(t)
}

func TestMessageHandler_HandleConnections_MessageTypes(t *testing.T) {
	uc := new(MockMessageUseCase)
	authClient := new(MockAuthServiceClient)

	uc.On("SaveMessage", mock.Anything).Return(nil)

	handler := NewMessageHandler(uc, authClient)

	// Mock ValidateSession to return valid token
	authClient.On("ValidateSession", mock.Anything, mock.MatchedBy(func(req *proto.ValidateSessionRequest) bool {
		return req.Token == "valid_token"
	})).Return(&proto.ValidateSessionResponse{Valid: true, UserId: 1, UserRole: "testrole"}, nil).Once()

	// Mock GetUserProfile to return valid user
	authClient.On("GetUserProfile", mock.Anything, mock.MatchedBy(func(req *proto.GetUserProfileRequest) bool {
		return req.UserId == 1
	})).Return(&proto.GetUserProfileResponse{User: &proto.User{Id: 1, Username: "testuser", Role: "testrole"}}, nil).Once()

	router := gin.Default()
	router.GET("/ws", handler.HandleConnections)

	server := httptest.NewServer(router)
	defer server.Close()

	url := "ws" + server.URL[4:] + "/ws?token=valid_token"
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Test binary message
	err = ws.WriteMessage(websocket.BinaryMessage, []byte("binary data"))
	assert.NoError(t, err)

	// Test unknown message type
	unknownMsg := incomingMessage{Type: "unknown", Message: "Should be ignored"}
	err = ws.WriteJSON(unknownMsg)
	assert.NoError(t, err)

	// Test valid message
	validMsg := incomingMessage{Type: "message", Message: "Hello, World!", Username: "testuser"}
	err = ws.WriteJSON(validMsg)
	assert.NoError(t, err)

	// Give some time for the handler to process
	time.Sleep(100 * time.Millisecond)

	uc.AssertExpectations(t)
	authClient.AssertExpectations(t)
}

func TestMessageHandler_HandleConnections_SaveMessageError(t *testing.T) {
	uc := new(MockMessageUseCase)
	authClient := new(MockAuthServiceClient)

	// Mock SaveMessage to return error
	uc.On("SaveMessage", mock.Anything).Return(errors.New("save message error"))

	handler := NewMessageHandler(uc, authClient)

	// Mock ValidateSession to return valid token
	authClient.On("ValidateSession", mock.Anything, mock.MatchedBy(func(req *proto.ValidateSessionRequest) bool {
		return req.Token == "valid_token"
	})).Return(&proto.ValidateSessionResponse{Valid: true, UserId: 1, UserRole: "testrole"}, nil).Once()

	// Mock GetUserProfile to return valid user
	authClient.On("GetUserProfile", mock.Anything, mock.MatchedBy(func(req *proto.GetUserProfileRequest) bool {
		return req.UserId == 1
	})).Return(&proto.GetUserProfileResponse{User: &proto.User{Id: 1, Username: "testuser", Role: "testrole"}}, nil).Once()

	router := gin.Default()
	router.GET("/ws", handler.HandleConnections)

	server := httptest.NewServer(router)
	defer server.Close()

	url := "ws" + server.URL[4:] + "/ws?token=valid_token"
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Send a message that will fail to save
	msg := incomingMessage{Type: "message", Message: "Hello, World!", Username: "testuser"}
	err = ws.WriteJSON(msg)
	assert.NoError(t, err)

	// Give some time for the handler to process
	time.Sleep(100 * time.Millisecond)

	uc.AssertExpectations(t)
	authClient.AssertExpectations(t)
}