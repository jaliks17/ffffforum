package mocks

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"backend/proto"
	"chat-service/internal/entity"
	"chat-service/internal/handler"
	"chat-service/internal/repository"
	"chat-service/internal/usecase"
	myWeb "chat-service/pkg/websocket"

	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type MessageIntegrationTestSuite struct {
	suite.Suite
	db        *sql.DB
	repo      repository.MessageRepository
	messageUC usecase.MessageUseCase
}

const (
	testDBHost     = "localhost"
	testDBPort     = "5432"
	testDBUser     = "user"
	testDBPassword = "555527"
	testDBName     = "database"
)

func (suite *MessageIntegrationTestSuite) SetupSuite() {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		testDBHost,
		testDBPort,
		testDBUser,
		testDBPassword,
		testDBName,
	)

	var err error
	suite.db, err = sql.Open("postgres", connStr)
	if err != nil {
		suite.T().Fatalf("Failed to connect to test database: %v", err)
	}

	err = suite.db.Ping()
	if err != nil {
		suite.T().Fatalf("Failed to ping test database: %v", err)
	}

	_, err = suite.db.Exec(`
        CREATE TABLE IF NOT EXISTS chat_messages (
            id SERIAL PRIMARY KEY,
            author_id INT NOT NULL,
            author_name VARCHAR(255) NOT NULL,
            content TEXT NOT NULL
        )
    `)
	if err != nil {
		suite.T().Fatalf("Failed to create test table: %v", err)
	}

	suite.repo = repository.NewMessageRepository(suite.db)
	suite.messageUC = usecase.NewMessageUseCase(suite.repo)
}

func (suite *MessageIntegrationTestSuite) TearDownSuite() {
	_, err := suite.db.Exec("DROP TABLE IF EXISTS chat_messages")
	if err != nil {
		suite.T().Logf("Warning: failed to drop test table: %v", err)
	}

	err = suite.db.Close()
	if err != nil {
		suite.T().Logf("Warning: failed to close database connection: %v", err)
	}
}

func (suite *MessageIntegrationTestSuite) SetupTest() {
	_, err := suite.db.Exec("TRUNCATE TABLE chat_messages RESTART IDENTITY CASCADE")
	if err != nil {
		suite.T().Fatalf("Failed to truncate test table: %v", err)
	}
}

func TestMessageIntegrationTestSuite(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests. Set RUN_INTEGRATION_TESTS=true to run them.")
	}
	suite.Run(t, new(MessageIntegrationTestSuite))
}

func (suite *MessageIntegrationTestSuite) TestSaveMessage() {
	tests := []struct {
		name        string
		message     entity.Message
		expectError bool
	}{
		{
			name: "successful save",
			message: entity.Message{
				UserID:   1,
				Username: "user1",
				Message:    "Hello",
			},
			expectError: false,
		},
		{
			name: "empty message",
			message: entity.Message{
				UserID:   1,
				Username: "user1",
				Message:    "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {

			_, err := suite.db.Exec("DELETE FROM chat_messages")
			assert.NoError(suite.T(), err)

			err = suite.messageUC.SaveMessage(&tt.message)
			if tt.expectError {
				assert.Error(suite.T(), err)
			} else {
				assert.NoError(suite.T(), err)
			}

			messages, err := suite.repo.GetMessages()
			assert.NoError(suite.T(), err)
			assert.Len(suite.T(), messages, 1, "Должно быть ровно одно сообщение в базе")
			assert.Equal(suite.T(), tt.message.Username, messages[0].Username)
			assert.Equal(suite.T(), tt.message.Message, messages[0].Message)
		})
	}
}
func (suite *MessageIntegrationTestSuite) TestGetMessages() {

	messagesToSave := []entity.Message{
		{UserID: 1, Username: "user1", Message: "Message 1"},
		{UserID: 2, Username: "user2", Message: "Message 2"},
	}

	for _, msg := range messagesToSave {
		err := suite.repo.SaveMessage(&msg)
		assert.NoError(suite.T(), err)
	}

	messages, err := suite.messageUC.GetMessages()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), messages, len(messagesToSave))

	for i, msg := range messages {
		assert.Equal(suite.T(), messagesToSave[i].Username, msg.Username)
		assert.Equal(suite.T(), messagesToSave[i].Message, msg.Message)
		assert.NotZero(suite.T(), msg.ID)
	}
}

func (suite *MessageIntegrationTestSuite) TestMessageFlow() {

	testMsg := entity.Message{
		UserID:   1,
		Username: "testuser",
		Message:    "Integration test message",
	}

	err := suite.messageUC.SaveMessage(&testMsg)
	assert.NoError(suite.T(), err)

	messages, err := suite.messageUC.GetMessages()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), messages, 1)
	assert.Equal(suite.T(), testMsg.Username, messages[0].Username)
	assert.Equal(suite.T(), testMsg.Message, messages[0].Message)
}

func TestGetMessages_EdgeCases(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewMessageRepository(db)

	t.Run("empty result", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, author_id, author_name, content FROM chat_messages").
			WillReturnRows(sqlmock.NewRows([]string{"id", "author_id", "author_name", "content"}))

		messages, err := repo.GetMessages()
		require.NoError(t, err)
		require.Empty(t, messages)
	})

	t.Run("scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "author_id"}).
			AddRow(1, "user1")
		mock.ExpectQuery("SELECT id, author_id, author_name, content FROM chat_messages").
			WillReturnRows(rows)

		_, err := repo.GetMessages()
		require.Error(t, err)
	})
}

type mockAuthServiceClient struct {
	mock.Mock
}

func (m *mockAuthServiceClient) GetUserProfile(ctx context.Context, in *proto.GetUserProfileRequest, opts ...grpc.CallOption) (*proto.GetUserProfileResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.GetUserProfileResponse), args.Error(1)
}

func (m *mockAuthServiceClient) Login(ctx context.Context, in *proto.LoginRequest, opts ...grpc.CallOption) (*proto.TokenResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.TokenResponse), args.Error(1)
}

func (m *mockAuthServiceClient) Register(ctx context.Context, in *proto.RegisterRequest, opts ...grpc.CallOption) (*proto.UserResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.UserResponse), args.Error(1)
}

func (m *mockAuthServiceClient) ValidateToken(ctx context.Context, in *proto.ValidateTokenRequest, opts ...grpc.CallOption) (*proto.ValidateSessionResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.ValidateSessionResponse), args.Error(1)
}

func (m *mockAuthServiceClient) RefreshToken(ctx context.Context, in *proto.RefreshTokenRequest, opts ...grpc.CallOption) (*proto.TokenResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.TokenResponse), args.Error(1)
}

func (m *mockAuthServiceClient) Logout(ctx context.Context, in *proto.LogoutRequest, opts ...grpc.CallOption) (*proto.SuccessResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.SuccessResponse), args.Error(1)
}

func (m *mockAuthServiceClient) SignIn(ctx context.Context, in *proto.SignInRequest, opts ...grpc.CallOption) (*proto.SignInResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.SignInResponse), args.Error(1)
}

func (m *mockAuthServiceClient) SignUp(ctx context.Context, in *proto.SignUpRequest, opts ...grpc.CallOption) (*proto.SignUpResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.SignUpResponse), args.Error(1)
}

func (m *mockAuthServiceClient) ValidateSession(ctx context.Context, in *proto.ValidateSessionRequest, opts ...grpc.CallOption) (*proto.ValidateSessionResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.ValidateSessionResponse), args.Error(1)
}

func TestMessageHandler(t *testing.T) {
	mockUC := &mockMessageUseCase{
		saveFunc: func(msg *entity.Message) error {
			return nil
		},
		getMessagesFunc: func() ([]entity.Message, error) {
			return []entity.Message{
				{ID: 1, UserID: 1, Username: "user1", Message: "Hello"},
				{ID: 2, UserID: 2, Username: "user2", Message: "Hi there"},
			}, nil
		},
	}

	authClient := new(mockAuthServiceClient)
	h := handler.NewMessageHandler(mockUC, authClient)

	t.Run("GetMessages success", func(t *testing.T) {
		router := gin.Default()
		router.GET("/messages", h.GetMessages)

		req, _ := http.NewRequest("GET", "/messages", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []entity.Message
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response, 2)
	})

	t.Run("GetMessages database error", func(t *testing.T) {
		errorUC := &mockMessageUseCase{
			getMessagesFunc: func() ([]entity.Message, error) {
				return nil, errors.New("database error")
			},
		}

		errorHandler := handler.NewMessageHandler(errorUC, authClient)

		router := gin.Default()
		router.GET("/messages", errorHandler.GetMessages)

		req, _ := http.NewRequest("GET", "/messages", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("HandleConnections websocket upgrade", func(t *testing.T) {

		myWeb.Clients = make(map[*websocket.Conn]int)

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			c, _ := gin.CreateTestContext(w)
			c.Request = r
			h.HandleConnections(c)
		}))
		defer s.Close()

		u := "ws" + strings.TrimPrefix(s.URL, "http")
		ws, _, err := websocket.DefaultDialer.Dial(u+"/ws", nil)
		require.NoError(t, err)
		defer ws.Close()

		assert.Equal(t, 1, len(myWeb.Clients))

		testMsg := entity.Message{UserID: 1, Username: "test", Message: "hello"}
		err = ws.WriteJSON(testMsg)
		require.NoError(t, err)

		assert.Equal(t, 1, mockUC.saveCount)
	})

	t.Run("HandleMessages broadcast", func(t *testing.T) {

		myWeb.Clients = make(map[*websocket.Conn]int)
		myWeb.Broadcast = make(chan entity.Message)

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, _ := gin.CreateTestContext(w)
			c.Request = r
			h.HandleConnections(c)
		}))
		defer s.Close()

		u := "ws" + strings.TrimPrefix(s.URL, "http")
		ws1, _, err := websocket.DefaultDialer.Dial(u+"/ws", nil)
		require.NoError(t, err)
		defer ws1.Close()

		ws2, _, err := websocket.DefaultDialer.Dial(u+"/ws", nil)
		require.NoError(t, err)
		defer ws2.Close()

		go h.HandleMessages()

		broadcastMsg := entity.Message{UserID: 0, Username: "system", Message: "broadcast"}
		myWeb.Broadcast <- broadcastMsg

		var msg1, msg2 entity.Message
		err = ws1.ReadJSON(&msg1)
		require.NoError(t, err)
		assert.Equal(t, broadcastMsg.Message, msg1.Message)

		err = ws2.ReadJSON(&msg2)
		require.NoError(t, err)
		assert.Equal(t, broadcastMsg.Message, msg2.Message)
	})
}

type mockMessageUseCase struct {
	usecase.MessageUseCase
	saveFunc        func(*entity.Message) error
	getMessagesFunc func() ([]entity.Message, error)
	saveCount       int
}

func (m *mockMessageUseCase) SaveMessage(msg *entity.Message) error {
	m.saveCount++
	if m.saveFunc != nil {
		return m.saveFunc(msg)
	}
	return nil
}

func (m *mockMessageUseCase) GetMessages() ([]entity.Message, error) {
	if m.getMessagesFunc != nil {
		return m.getMessagesFunc()
	}
	return nil, nil
}