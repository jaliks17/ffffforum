package usecase

import (
	"context"
	"database/sql"
	"errors"
	"time"

	pb "backend/proto"
	"forum-service/internal/entity"
	"forum-service/internal/repository"
	"forum-service/pkg/logger"
)

type PostUsecase struct {
	postRepo   repository.PostRepository
	authClient pb.AuthServiceClient
	logger     *logger.Logger
}
type PostUsecaseInterface interface {
	CreatePost(ctx context.Context, token, title, content string) (*entity.Post, error)
	GetPosts(ctx context.Context) ([]*entity.Post, map[int]string, error)
	DeletePost(ctx context.Context, token string, postID int64) error
	UpdatePost(ctx context.Context, token string, postID int64, title, content string) (*entity.Post, error)
}

func NewPostUsecase(
	postRepo repository.PostRepository,
	authClient pb.AuthServiceClient,
	logger *logger.Logger,
) *PostUsecase {
	return &PostUsecase{
		postRepo:   postRepo,
		authClient: authClient,
		logger:     logger,
	}
}

func (uc *PostUsecase) CreatePost(ctx context.Context, token string, title, content string) (*entity.Post, error) {

	validateResp, err := uc.authClient.ValidateToken(ctx, &pb.ValidateTokenRequest{Token: token})
	if err != nil {
		return nil, err
	}
	if !validateResp.Valid {
		return nil, errors.New("invalid token")
	}
	userID := validateResp.UserId

	post := &entity.Post{
		Title:     title,
		Content:   content,
		AuthorID:  userID,
		CreatedAt: time.Now(),
	}

	id, err := uc.postRepo.CreatePost(ctx, post)
	if err != nil {
		return nil, err
	}

	post.ID = id
	return post, nil
}

func (uc *PostUsecase) GetPosts(ctx context.Context) ([]*entity.Post, map[int]string, error) {
	posts, err := uc.postRepo.GetPosts(ctx)
	if err != nil {
		return nil, nil, err
	}

	authorNames := make(map[int]string)

	authorIDs := make([]int64, 0, len(posts))
	for _, post := range posts {
		authorIDs = append(authorIDs, post.AuthorID)
	}

	for _, authorID := range authorIDs {
		userResponse, err := uc.authClient.GetUserProfile(ctx, &pb.GetUserProfileRequest{
			UserId: authorID,
		})

		if err == nil && userResponse.User != nil {
			authorNames[int(authorID)] = userResponse.User.Username
		}
	}

	for _, post := range posts {
		if _, exists := authorNames[int(post.AuthorID)]; !exists {
			authorNames[int(post.AuthorID)] = "Unknown"
		}
	}

	return posts, authorNames, nil
}
func (uc *PostUsecase) DeletePost(ctx context.Context, token string, postID int64) error {
	validateResp, err := uc.authClient.ValidateToken(ctx, &pb.ValidateTokenRequest{Token: token})
	if err != nil {
		return err
	}
	if !validateResp.Valid {
		return errors.New("invalid token")
	}

	err = uc.postRepo.DeletePost(
		ctx,
		postID,
		validateResp.UserId,
		validateResp.UserRole,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return errors.New("post not found")
		case errors.Is(err, repository.ErrPermissionDenied):
			return errors.New("permission denied")
		default:
			return err
		}
	}

	return nil
}

func (uc *PostUsecase) UpdatePost(
	ctx context.Context,
	token string,
	postID int64,
	title,
	content string,
) (*entity.Post, error) {
	validateResp, err := uc.authClient.ValidateToken(ctx, &pb.ValidateTokenRequest{Token: token})
	if err != nil {
		return nil, err
	}
	if !validateResp.Valid {
		return nil, errors.New("invalid token")
	}

	updatedPost, err := uc.postRepo.UpdatePost(
		ctx,
		postID,
		validateResp.UserId,
		validateResp.UserRole,
		title,
		content,
	)

	return updatedPost, err
}