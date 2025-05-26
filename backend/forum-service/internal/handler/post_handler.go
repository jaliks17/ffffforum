package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "forum-service/docs"
	"forum-service/internal/repository"
	"forum-service/internal/usecase"
	"forum-service/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PostHandler struct {
	uc     usecase.PostUsecaseInterface
	logger *logger.Logger
}

func NewPostHandler(uc usecase.PostUsecaseInterface, logger *logger.Logger) *PostHandler {
	return &PostHandler{uc: uc, logger: logger}
}

// CreatePost godoc
// @Summary Создать новый пост
// @Description Создает новый пост в системе
// @Tags Посты
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer токен"
// @Param request body entity.Post true "Данные поста"
// @Success 201 {object} entity.Post
// @Failure 400 {object} entity.ErrorResponse
// @Failure 401 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/posts [post]

func (h *PostHandler) CreatePost(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	var request struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	post, err := h.uc.CreatePost(ctx.Request.Context(), token, request.Title, request.Content)
	if err != nil {
		h.logger.Error("Failed to create post", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"id":      post.ID,
		"message": "Post created successfully",
		"post":    post,
	})
}

// GetPosts godoc
// @Summary Get all posts
// @Description Get list of all forum posts with pagination
// @Tags posts
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Posts per page" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/posts [get]
func (h *PostHandler) GetPosts(c *gin.Context) {
	posts, authorNames, err := h.uc.GetPosts(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get posts", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get posts",
			"details": err.Error(),
		})
		return
	}

	response := make([]gin.H, 0, len(posts))
	for _, post := range posts {
		response = append(response, gin.H{
			"id":          post.ID,
			"title":       post.Title,
			"content":     post.Content,
			"author_id":   post.AuthorID,
			"author_name": authorNames[int(post.AuthorID)],
			"created_at":  post.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// DeletePost godoc
// @Summary Delete a post
// @Description Delete a forum post by ID (only author or admin can delete)
// @Tags posts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param id path int true "Post ID"
// @Success 200 {object} entity.SuccessResponse
// @Failure 400 {object} entity.ErrorResponse
// @Failure 401 {object} entity.ErrorResponse
// @Failure 403 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/posts/{id} [delete]
func (h *PostHandler) DeletePost(ctx *gin.Context) {
	postID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	h.logger.Debug("Attempting to delete post",
		zap.Int64("post_id", postID),
		zap.String("token", token),
	)

	if err := h.uc.DeletePost(ctx.Request.Context(), token, postID); err != nil {
		h.logger.Error("Failed to delete post", err)

		switch {
		case errors.Is(err, repository.ErrPostNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		case errors.Is(err, repository.ErrPermissionDenied):
			ctx.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

// UpdatePost godoc
// @Summary Update a post
// @Description Update an existing forum post (only author can update)
// @Tags posts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param id path int true "Post ID"
// @Param request body entity.Post true "Update data"
// @Success 200 {object} entity.Post
// @Failure 400 {object} entity.ErrorResponse
// @Failure 401 {object} entity.ErrorResponse
// @Failure 403 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/posts/{id} [put]
func (h *PostHandler) UpdatePost(ctx *gin.Context) {
	postID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	var request struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	updatedPost, err := h.uc.UpdatePost(ctx.Request.Context(), token, postID, request.Title, request.Content)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrPostNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		case errors.Is(err, repository.ErrPermissionDenied):
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		default:
			h.logger.Error("Failed to update post", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Post updated successfully",
		"post":    updatedPost,
	})
}