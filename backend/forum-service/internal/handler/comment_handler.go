// internal/handler/comment_handler.go
package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	pb "github.com/jaliks17/ffffforum/backend/proto"

	"github.com/jaliks17/ffffforum/backend/forum-service/internal/entity"
	"github.com/jaliks17/ffffforum/backend/forum-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	commentUC usecase.CommentUseCaseInterface
}

func NewCommentHandler(commentUC usecase.CommentUseCaseInterface) *CommentHandler {
	return &CommentHandler{commentUC: commentUC}
}

// CreateComment godoc
// @Summary Create a new comment
// @Description Create a new comment for a specific post
// @Tags comments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param id path int true "Post ID"
// @Param request body entity.Comment true "Comment data"
// @Success 201 {object} entity.Comment
// @Failure 400 {object} entity.ErrorResponse
// @Failure 401 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/posts/{id}/comments [post]
func (h *CommentHandler) CreateComment(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	authResponse, err := h.commentUC.GetAuthClient().ValidateToken(c.Request.Context(), &pb.ValidateTokenRequest{
		Token: token,
	})
	if err != nil || authResponse == nil || !authResponse.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	var request struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userResponse, err := h.commentUC.GetAuthClient().GetUserProfile(c.Request.Context(), &pb.GetUserProfileRequest{
		UserId: authResponse.UserId,
	})

	comment := entity.Comment{
		Content:    request.Content,
		AuthorID:   authResponse.UserId,
		AuthorName: "Unknown",
		PostID:     postID,
	}

	if err == nil && userResponse != nil && userResponse.User != nil {
		comment.AuthorName = userResponse.User.Username
	}

	if err := h.commentUC.CreateComment(c.Request.Context(), &comment); err != nil {
		log.Printf("Error creating comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          comment.ID,
		"content":     comment.Content,
		"author_id":   comment.AuthorID,
		"post_id":     comment.PostID,
		"author_name": comment.AuthorName,
	})
}

// GetCommentsByPostID godoc
// @Summary Get comments for a post
// @Description Get all comments for a specific post
// @Tags comments
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} []entity.Comment
// @Failure 400 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/posts/{id}/comments [get]
func (h *CommentHandler) GetCommentsByPostID(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Printf("Invalid post ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	comments, err := h.commentUC.GetCommentsByPostID(c.Request.Context(), postID)
	if err != nil {
		log.Printf("Error getting comments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get comments",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
	})
}

// DeleteComment godoc
// @Summary Delete a comment
// @Description Delete a comment by ID (only author or admin can delete)
// @Tags comments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param id path int true "Comment ID"
// @Success 200 {object} entity.SuccessResponse
// @Failure 400 {object} entity.ErrorResponse
// @Failure 401 {object} entity.ErrorResponse
// @Failure 403 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/comments/{id} [delete]
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	authResponse, err := h.commentUC.GetAuthClient().ValidateToken(c.Request.Context(), &pb.ValidateTokenRequest{
		Token: token,
	})
	if err != nil || authResponse == nil || !authResponse.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	if err := h.commentUC.DeleteComment(c.Request.Context(), commentID, authResponse.UserId); err != nil {
		switch {
		case errors.Is(err, usecase.ErrCommentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		case errors.Is(err, usecase.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this comment"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}

// Create godoc
// @Summary Create a new comment
// @Description Create a new comment for a post
// @Tags comments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param request body entity.Comment true "Comment data"
// @Success 201 {object} entity.Comment
// @Failure 400 {object} entity.ErrorResponse
// @Failure 401 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/comments [post]
func (h *CommentHandler) Create(ctx context.Context, comment *entity.Comment) error {
	return h.commentUC.CreateComment(ctx, comment)
}

// Get godoc
// @Summary Get a comment
// @Description Get a comment by ID
// @Tags comments
// @Accept json
// @Produce json
// @Param id path int true "Comment ID"
// @Success 200 {object} entity.Comment
// @Failure 400 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/comments/{id} [get]
func (h *CommentHandler) Get(ctx context.Context, id int64) (*entity.Comment, error) {
	return h.commentUC.GetComment(ctx, id)
}

// LikeComment godoc
// @Summary Like a comment
// @Description Like or unlike a comment
// @Tags comments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param id path int true "Comment ID"
// @Success 200 {object} map[string]interface{} "likes, is_liked"
// @Failure 400 {object} entity.ErrorResponse
// @Failure 401 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/comments/{id}/like [post]
func (h *CommentHandler) LikeComment(c *gin.Context) {
	// Placeholder implementation
	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	// TODO: Implement like/unlike logic in usecase
	// For now, just return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"comment_id": commentID,
		"likes":      1,    // Placeholder
		"is_liked":   true, // Placeholder
	})
}