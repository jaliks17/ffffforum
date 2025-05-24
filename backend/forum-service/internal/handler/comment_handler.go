package handler

import (
	"context"
	"forum-service/internal/entity"
	"forum-service/internal/usecase"
	"log"
	"net/http"
	"strconv"
	"strings"

	pb "backend/proto"

	"github.com/gin-gonic/gin"
)

type CommentUseCase interface {
	CreateComment(ctx context.Context, comment *entity.Comment) error
	GetComment(ctx context.Context, id int64) (*entity.Comment, error)
	DeleteComment(ctx context.Context, id int64, userID int64) error
}

type CommentHandler struct {
	usecase    CommentUseCase
	authClient pb.AuthServiceClient
}

func NewCommentHandler(uc CommentUseCase, authClient pb.AuthServiceClient) *CommentHandler {
	return &CommentHandler{
		usecase:    uc,
		authClient: authClient,
	}
}

func (h *CommentHandler) Create(ctx context.Context, comment *entity.Comment) error {
	return h.usecase.CreateComment(ctx, comment)
}

func (h *CommentHandler) Get(ctx context.Context, id int64) (*entity.Comment, error) {
	return h.usecase.GetComment(ctx, id)
}

func (h *CommentHandler) Delete(ctx context.Context, id int64, userID int64) error {
	return h.usecase.DeleteComment(ctx, id, userID)
}

// DeleteComment godoc
// @Summary Delete a comment
// @Description Delete a comment by its ID
// @Tags comments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param id path int true "Comment ID"
// @Success 200 {object} map[string]string "message"
// @Failure 400 {object} entity.ErrorResponse
// @Failure 401 {object} entity.ErrorResponse
// @Failure 403 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/comments/{id} [delete]
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	authResponse, err := h.authClient.ValidateToken(c.Request.Context(), &pb.ValidateTokenRequest{
		Token: token,
	})
	if err != nil || authResponse == nil || !authResponse.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	// Call the usecase method to delete the comment
	if err := h.usecase.DeleteComment(c.Request.Context(), commentID, authResponse.UserId); err != nil {
		log.Printf("Error deleting comment %d for user %d: %v", commentID, authResponse.UserId, err)
		// Handle specific errors from usecase (e.g., not found, forbidden)
		if err == usecase.ErrCommentNotFound { // Assuming ErrCommentNotFound is defined in usecase
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		} else if err == usecase.ErrForbidden { // Assuming ErrForbidden is defined in usecase
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this comment"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}

func (h *CommentHandler) CreateComment(c *gin.Context) {
	var comment entity.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.usecase.CreateComment(c.Request.Context(), &comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

func (h *CommentHandler) GetComment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	comment, err := h.usecase.GetComment(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrCommentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get comment"})
		}
		return
	}

	c.JSON(http.StatusOK, comment)
}

func (h *CommentHandler) RegisterRoutes(router *gin.Engine) {
	comments := router.Group("/api/comments")
	{
		comments.POST("", h.CreateComment)
		comments.GET("/:id", h.GetComment)
		comments.DELETE("/:id", h.DeleteComment)
	}
}