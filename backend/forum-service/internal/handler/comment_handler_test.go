package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/jaliks17/ffffforum/backend/forum-service/internal/entity"
	"github.com/jaliks17/ffffforum/backend/forum-service/internal/usecase"

	pb "github.com/jaliks17/ffffforum/backend/proto"

	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type MockAuthServiceClient struct {
	mock.Mock
}

func (m *MockAuthServiceClient) ValidateToken(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ValidateSessionResponse), args.Error(1)
}

func (m *MockAuthServiceClient) GetUserProfile(ctx context.Context, in *pb.GetUserProfileRequest, opts ...grpc.CallOption) (*pb.GetUserProfileResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetUserProfileResponse), args.Error(1)
}

func (m *MockAuthServiceClient) Login(ctx context.Context, in *pb.LoginRequest, opts ...grpc.CallOption) (*pb.TokenResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.TokenResponse), args.Error(1)
}

func (m *MockAuthServiceClient) Register(ctx context.Context, in *pb.RegisterRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UserResponse), args.Error(1)
}

func (m *MockAuthServiceClient) Logout(ctx context.Context, in *pb.LogoutRequest, opts ...grpc.CallOption) (*pb.SuccessResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.SuccessResponse), args.Error(1)
}

func (m *MockAuthServiceClient) RefreshToken(ctx context.Context, in *pb.RefreshTokenRequest, opts ...grpc.CallOption) (*pb.TokenResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.TokenResponse), args.Error(1)
}

func (m *MockAuthServiceClient) SignIn(ctx context.Context, in *pb.SignInRequest, opts ...grpc.CallOption) (*pb.SignInResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.SignInResponse), args.Error(1)
}

func (m *MockAuthServiceClient) SignUp(ctx context.Context, in *pb.SignUpRequest, opts ...grpc.CallOption) (*pb.SignUpResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.SignUpResponse), args.Error(1)
}

func (m *MockAuthServiceClient) ValidateSession(ctx context.Context, in *pb.ValidateSessionRequest, opts ...grpc.CallOption) (*pb.ValidateSessionResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ValidateSessionResponse), args.Error(1)
}

type MockCommentUseCase struct {
	mock.Mock
	usecase.CommentUseCase
}

func (m *MockCommentUseCase) CreateComment(ctx context.Context, comment *entity.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentUseCase) GetComment(ctx context.Context, id int64) (*entity.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Comment), args.Error(1)
}

func (m *MockCommentUseCase) DeleteComment(ctx context.Context, id int64, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockCommentUseCase) GetCommentsByPostID(ctx context.Context, postID int64) ([]entity.Comment, error) {
	args := m.Called(ctx, postID)
	return args.Get(0).([]entity.Comment), args.Error(1)
}

func (m *MockCommentUseCase) GetAuthClient() pb.AuthServiceClient {
	args := m.Called()
	return args.Get(0).(pb.AuthServiceClient)
}

func TestCommentHandler_DeleteComment(t *testing.T) {
	tests := []struct {
		name           string
		commentID      string
		authHeader     string
		mockAuthResp   *pb.ValidateSessionResponse
		mockAuthErr    error
		mockDeleteErr  error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "successful deletion",
			commentID:  "1",
			authHeader: "Bearer valid-token",
			mockAuthResp: &pb.ValidateSessionResponse{
				Valid:  true,
				UserId: 1,
			},
			expectedStatus: 200,
			expectedBody: map[string]interface{}{
				"message": "Comment deleted successfully",
			},
		},
		{
			name:           "missing auth header",
			commentID:      "1",
			authHeader:     "",
			expectedStatus: 401,
			expectedBody: map[string]interface{}{
				"error": "Authorization header is required",
			},
		},
		{
			name:       "invalid token",
			commentID:  "1",
			authHeader: "Bearer invalid-token",
			mockAuthErr: errors.New("invalid token"),
			expectedStatus: 401,
			expectedBody: map[string]interface{}{
				"error": "Invalid token",
			},
		},
		{
			name:       "comment not found",
			commentID:  "999",
			authHeader: "Bearer valid-token",
			mockAuthResp: &pb.ValidateSessionResponse{
				Valid:  true,
				UserId: 1,
			},
			mockDeleteErr: usecase.ErrCommentNotFound,
			expectedStatus: 404,
			expectedBody: map[string]interface{}{
				"error": "Comment not found",
			},
		},
		{
			name:       "permission denied",
			commentID:  "1",
			authHeader: "Bearer valid-token",
			mockAuthResp: &pb.ValidateSessionResponse{
				Valid:  true,
				UserId: 1,
			},
			mockDeleteErr: usecase.ErrForbidden,
			expectedStatus: 403,
			expectedBody: map[string]interface{}{
				"error": "You do not have permission to delete this comment",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("DELETE", "/comments/"+tt.commentID, nil)
			if tt.authHeader != "" {
				c.Request.Header.Set("Authorization", tt.authHeader)
			}
			c.Params = []gin.Param{{Key: "id", Value: tt.commentID}}

			mockUC := new(MockCommentUseCase)
			h := NewCommentHandler(mockUC)

			// Setup mock expectations
			mockAuthClient := new(MockAuthServiceClient)
			mockUC.On("GetAuthClient").Return(mockAuthClient).Maybe()

			// Mock ValidateToken call
			if tt.authHeader != "" {
				mockAuthClient.On("ValidateToken", mock.Anything, mock.Anything, mock.Anything).Return(tt.mockAuthResp, tt.mockAuthErr).Once()
			}

			// Mock DeleteComment call only if auth is expected to succeed
			if tt.authHeader != "" && tt.mockAuthResp != nil && tt.mockAuthResp.Valid {
				mockUC.On("DeleteComment", mock.Anything, mock.Anything, tt.mockAuthResp.UserId).Return(tt.mockDeleteErr)
			}

			// Execute
			h.DeleteComment(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			// assert.Equal(t, tt.expectedBody, w.Body) // Commented out due to inconsistency with Gin's response body type

			// More robust body assertion
			if tt.expectedBody != nil {
				expectedJSON, _ := json.Marshal(tt.expectedBody)
				assert.JSONEq(t, string(expectedJSON), w.Body.String())
			}

			mockUC.AssertExpectations(t)
			mockAuthClient.AssertExpectations(t)
		})
	}
}

func TestCommentHandler_CreateComment(t *testing.T) {
	tests := []struct {
		name           string
		postID         string
		authHeader     string
		reqBody        string
		mockAuthResp   *pb.ValidateSessionResponse
		mockAuthErr    error
		mockUserProfileResp *pb.GetUserProfileResponse
		mockUserProfileErr error
		mockCreateErr  error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "successful creation",
			postID:     "1",
			authHeader: "Bearer valid-token",
			reqBody:    `{"content":"test comment"}`,
			mockAuthResp: &pb.ValidateSessionResponse{
				Valid:  true,
				UserId: 42,
			},
			mockUserProfileResp: &pb.GetUserProfileResponse{User: &pb.User{Username: "alice"}},
			expectedStatus: 201,
			expectedBody: map[string]interface{}{
				"id":          float64(0), // ID is assigned by repository, so mock 0
				"content":     "test comment",
				"author_id":   float64(42),
				"post_id":     float64(1),
				"author_name": "alice",
			},
		},
		{
			name:           "missing auth header",
			postID:         "1",
			authHeader:     "",
			reqBody:        `{"content":"test comment"}`,
			expectedStatus: 401,
			expectedBody: map[string]interface{}{
				"error": "Authorization header is required",
			},
		},
		{
			name:       "invalid token",
			postID:     "1",
			authHeader: "Bearer invalid-token",
			reqBody:    `{"content":"test comment"}`,
			mockAuthErr: errors.New("invalid token"),
			expectedStatus: 401,
			expectedBody: map[string]interface{}{
				"error": "Invalid token",
			},
		},
		{
			name:           "invalid post id",
			postID:         "invalid",
			authHeader:     "Bearer valid-token",
			reqBody:        `{"content":"test comment"}`,
			mockAuthResp: &pb.ValidateSessionResponse{
				Valid:  true,
				UserId: 42,
			},
			mockUserProfileResp: &pb.GetUserProfileResponse{User: &pb.User{Username: "alice"}},
			expectedStatus: 400,
			expectedBody: map[string]interface{}{
				"error": "invalid post id",
			},
		},
		{
			name:       "invalid request body",
			postID:     "1",
			authHeader: "Bearer valid-token",
			reqBody:    `{"invalid_field":"test"}`,
			mockAuthResp: &pb.ValidateSessionResponse{
				Valid:  true,
				UserId: 42,
			},
			mockUserProfileResp: &pb.GetUserProfileResponse{User: &pb.User{Username: "alice"}},
			expectedStatus: 400,
			expectedBody: map[string]interface{}{
				"error": "Invalid request body",
			},
		},
		{
			name:       "usecase error",
			postID:     "1",
			authHeader: "Bearer valid-token",	
			reqBody:    `{"content":"test comment"}`,
			mockAuthResp: &pb.ValidateSessionResponse{
				Valid:  true,
				UserId: 42,
			},
			mockUserProfileResp: &pb.GetUserProfileResponse{User: &pb.User{Username: "alice"}},
			mockCreateErr: errors.New("failed to create comment"),
			expectedStatus: 500,
			expectedBody: map[string]interface{}{
				"error": "Failed to create comment",
			},
		},
		{
			name:       "GetUserProfile error",
			postID:     "1",
			authHeader: "Bearer valid-token",	
			reqBody:    `{"content":"test comment"}`,
			mockAuthResp: &pb.ValidateSessionResponse{
				Valid:  true,
				UserId: 42,
			},
			mockUserProfileErr: errors.New("grpc error"),
			expectedStatus: 201, // Still success, but author name will be "Unknown"
			expectedBody: map[string]interface{}{
				"id":          float64(0), // ID is assigned by repository, so mock 0
				"content":     "test comment",
				"author_id":   float64(42),
				"post_id":     float64(1),
				"author_name": "Unknown", // Should be Unknown due to error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			
			// Construct request with post ID in path
			c.Request = httptest.NewRequest("POST", "/posts/"+tt.postID+"/comments", bytes.NewBufferString(tt.reqBody))
			c.Request.Header.Set("Content-Type", "application/json")
			
			if tt.authHeader != "" {
				c.Request.Header.Set("Authorization", tt.authHeader)
			}
			c.Params = []gin.Param{{Key: "id", Value: tt.postID}}

			mockUC := new(MockCommentUseCase)
			h := NewCommentHandler(mockUC)

			// Setup mock expectations
			mockAuthClient := new(MockAuthServiceClient)
			mockUC.On("GetAuthClient").Return(mockAuthClient).Maybe()

			// Mock ValidateToken call if auth header is present
			if tt.authHeader != "" {
				mockAuthClient.On("ValidateToken", mock.Anything, mock.Anything, mock.Anything).Return(tt.mockAuthResp, tt.mockAuthErr).Maybe()
			}

			// Mock GetUserProfile call if auth is expected to succeed
			if tt.authHeader != "" && tt.mockAuthResp != nil && tt.mockAuthResp.Valid {
				mockAuthClient.On("GetUserProfile", mock.Anything, mock.Anything, mock.Anything).Return(tt.mockUserProfileResp, tt.mockUserProfileErr).Maybe()
			}

			// Mock CreateComment call only if previous steps succeed or if GetUserProfile returns error
			if tt.authHeader != "" && tt.mockAuthResp != nil && tt.mockAuthResp.Valid {
				expectedComment := &entity.Comment{
					Content:    "", // Will be dynamically set
					AuthorID:   tt.mockAuthResp.UserId,
					PostID:     func() int64 { // Handle invalid post ID case
						postIDInt, err := strconv.ParseInt(tt.postID, 10, 64)
						if err != nil { return 0 } 
						return postIDInt
					}(),
					AuthorName: "", // Will be dynamically set
				}
                // Dynamically parse content from reqBody for the expected comment mock
                var reqBodyMap map[string]string
                json.Unmarshal([]byte(tt.reqBody), &reqBodyMap)
                expectedComment.Content = reqBodyMap["content"]

                // Set AuthorName based on whether GetUserProfile is expected to succeed
                if tt.mockUserProfileErr == nil && tt.mockUserProfileResp != nil && tt.mockUserProfileResp.User != nil {
                    expectedComment.AuthorName = tt.mockUserProfileResp.User.Username
                } else {
                    expectedComment.AuthorName = "Unknown"
                }

				mockUC.On("CreateComment", mock.Anything, expectedComment).Return(tt.mockCreateErr).Maybe()
			}

			// Execute
			h.CreateComment(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// More robust body assertion
			if tt.expectedBody != nil {
                // For success case, dynamically set expected body values that depend on input/mocks
                if tt.name == "successful creation" || tt.name == "GetUserProfile error" {
                    var reqBodyMap map[string]string
                    json.Unmarshal([]byte(tt.reqBody), &reqBodyMap)
                    tt.expectedBody["content"] = reqBodyMap["content"]
                    tt.expectedBody["author_id"] = float64(tt.mockAuthResp.UserId)
                    if tt.name == "successful creation"{
                         tt.expectedBody["author_name"] = tt.mockUserProfileResp.User.Username
                    } else { // GetUserProfile error case
                         tt.expectedBody["author_name"] = "Unknown"
                    }
                   
                    postIDInt, err := strconv.ParseInt(tt.postID, 10, 64)
                     if err == nil { // Only set post_id if postID is valid
                         tt.expectedBody["post_id"] = float64(postIDInt)
                     }

                }
				expectedJSON, _ := json.Marshal(tt.expectedBody)
				assert.JSONEq(t, string(expectedJSON), w.Body.String())
			}

			mockUC.AssertExpectations(t)
			mockAuthClient.AssertExpectations(t)
		})
	}
}

func TestCommentHandler_Create(t *testing.T) {
	tests := []struct {
		name          string
		comment       *entity.Comment
		mockCreateErr error
		expectedErr   error
	}{
		{
			name:    "successful creation",
			comment: &entity.Comment{Content: "test comment"},
		},
		{
			name:          "creation error",
			comment:       &entity.Comment{Content: "test comment"},
			mockCreateErr: errors.New("database error"),
			expectedErr:   errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(MockCommentUseCase)
			h := NewCommentHandler(mockUC)

			mockUC.On("CreateComment", mock.Anything, tt.comment).Return(tt.mockCreateErr)

			err := h.Create(context.Background(), tt.comment)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommentHandler_Get(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		mockComment *entity.Comment
		mockErr     error
		expected    *entity.Comment
		expectedErr error
	}{
		{
			name:        "successful get",
			id:          1,
			mockComment: &entity.Comment{ID: 1, Content: "test comment"},
			expected:    &entity.Comment{ID: 1, Content: "test comment"},
		},
		{
			name:        "comment not found",
			id:          999,
			mockErr:     usecase.ErrCommentNotFound,
			expectedErr: usecase.ErrCommentNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(MockCommentUseCase)
			h := NewCommentHandler(mockUC)

			mockUC.On("GetComment", mock.Anything, tt.id).Return(tt.mockComment, tt.mockErr)

			comment, err := h.Get(context.Background(), tt.id)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, comment)
			}
		})
	}
}

func TestCommentHandler_GetCommentsByPostID(t *testing.T) {
	tests := []struct {
		name           string
		postID         string
		mockComments   []entity.Comment
		mockErr        error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "successful get",
			postID: "1",
			mockComments: []entity.Comment{
				{ID: 1, PostID: 1, Content: "Comment 1", AuthorID: 1, AuthorName: "user1"},
				{ID: 2, PostID: 1, Content: "Comment 2", AuthorID: 2, AuthorName: "user2"},
			},
			expectedStatus: 200,
			expectedBody: map[string]interface{}{
				"comments": []interface{}{
					map[string]interface{}{"id": float64(1), "post_id": float64(1), "content": "Comment 1", "author_id": float64(1), "author_name": "user1", "CreatedAt": "0001-01-01T00:00:00Z"},
					map[string]interface{}{"id": float64(2), "post_id": float64(1), "content": "Comment 2", "author_id": float64(2), "author_name": "user2", "CreatedAt": "0001-01-01T00:00:00Z"},
				},
			},
		},
		{
			name:           "no comments",
			postID:         "2",
			mockComments:   []entity.Comment{},
			expectedStatus: 200,
			expectedBody: map[string]interface{}{
				"comments": []interface{}{},
			},
		},
		{
			name:           "invalid post id",
			postID:         "invalid",
			expectedStatus: 400,
			expectedBody: map[string]interface{}{
				"error": "invalid post id",
			},
		},
		{
			name:           "usecase error",
			postID:         "3",
			mockErr:        errors.New("database error"),
			expectedStatus: 500,
			expectedBody: map[string]interface{}{
				"error":   "failed to get comments",
				"details": "database error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			
			// Construct request with post ID in path
			c.Request = httptest.NewRequest("GET", "/posts/"+tt.postID+"/comments", nil)
			c.Params = []gin.Param{{Key: "id", Value: tt.postID}}

			mockUC := new(MockCommentUseCase)
			h := NewCommentHandler(mockUC)

			// Setup mock expectations for usecase call (only if postID is valid)
			if tt.postID != "invalid" {
				postIDInt, err := strconv.ParseInt(tt.postID, 10, 64)
				if err == nil {
					mockUC.On("GetCommentsByPostID", mock.Anything, postIDInt).Return(tt.mockComments, tt.mockErr).Maybe()
				}
			}

			// Execute
			h.GetCommentsByPostID(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// More robust body assertion
			if tt.expectedBody != nil {
				expectedJSON, _ := json.Marshal(tt.expectedBody)
				assert.JSONEq(t, string(expectedJSON), w.Body.String())
			}

			mockUC.AssertExpectations(t)
		})
	}
}