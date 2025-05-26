package handler

import (
	"context"
	"errors"
	"forum-service/internal/entity"
	"forum-service/internal/usecase"
	"testing"

	pb "backend/proto"

	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
			if tt.authHeader != "" {
				mockUC.On("DeleteComment", mock.Anything, int64(1), tt.mockAuthResp.UserId).Return(tt.mockDeleteErr)
			}

			// Execute
			h.DeleteComment(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body)
			mockUC.AssertExpectations(t)
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