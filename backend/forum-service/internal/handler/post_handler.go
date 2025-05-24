package handler

import (
	"context"
	"errors"
	"forum-service/internal/entity"
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

	pb "backend/proto"
)

type AuthClient interface {
	ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateSessionResponse, error)
	GetUser(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.UserResponse, error)
}

type PostUseCase interface {
	CreatePost(ctx context.Context, post *entity.Post) error
	GetPostByID(ctx context.Context, id int64) (*entity.Post, error)
	DeletePost(ctx context.Context, id int64) error
}

type PostHandler struct {
	uc         usecase.PostUsecaseInterface
	logger     *logger.Logger
	AuthClient AuthClient
}

func NewPostHandler(uc usecase.PostUsecaseInterface, logger *logger.Logger, authClient AuthClient) *PostHandler {
	return &PostHandler{uc: uc, logger: logger, AuthClient: authClient}
}

func (h *PostHandler) Create(ctx context.Context, post *entity.Post) error {
	_, err := h.uc.CreatePost(ctx, post.Title, post.Content, "")
	return err
}

func (h *PostHandler) Get(ctx context.Context, id int64) (*entity.Post, error) {
	return h.uc.GetPostByID(ctx, id)
}

func (h *PostHandler) Delete(ctx context.Context, id int64) error {
	return h.uc.DeletePost(ctx, "", id)
}

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
			h.logger.Error("Failed to delete post", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

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

func (h *PostHandler) GetPostByID(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	post, err := h.uc.GetPostByID(c.Request.Context(), postID)
	if err != nil {
		if errors.Is(err, repository.ErrPostNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		} else {
			h.logger.Error("Failed to get post by ID", zap.Int64("post_id", postID), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to get post",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, post)
}

func (h *PostHandler) GetPostsByUserID(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	posts, err := h.uc.GetPostsByUserID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get posts by user ID", zap.Int64("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get posts by user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": posts,
	})
}

func (h *PostHandler) RegisterRoutes(router *gin.Engine) {
	posts := router.Group("/api/posts")
	{
		posts.POST("", h.CreatePost)
		posts.GET("", h.GetPosts)
		posts.GET("/:id", h.GetPostByID)
		posts.PUT("/:id", h.UpdatePost)
		posts.DELETE("/:id", h.DeletePost)
		posts.GET("/user/:id", h.GetPostsByUserID)
	}
}