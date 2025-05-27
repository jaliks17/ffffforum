package usecase

import (
	"context"
	"errors"

	pb "backend/proto"
	"forum-service/internal/entity"
	"forum-service/internal/repository"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrForbidden      = errors.New("forbidden")
)

type CommentUseCaseInterface interface {
	CreateComment(ctx context.Context, comment *entity.Comment) error
	GetComment(ctx context.Context, id int64) (*entity.Comment, error)
	DeleteComment(ctx context.Context, id int64, userID int64) error
	GetCommentsByPostID(ctx context.Context, postID int64) ([]entity.Comment, error)
	GetAuthClient() pb.AuthServiceClient
}

type CommentUseCase struct {
	CommentRepo repository.CommentRepository
	PostRepo    repository.PostRepository
	AuthClient  pb.AuthServiceClient
}

func NewCommentUseCase(
	commentRepo repository.CommentRepository,
	postRepo repository.PostRepository,
	authClient pb.AuthServiceClient,
) *CommentUseCase {
	return &CommentUseCase{
		CommentRepo: commentRepo,
		PostRepo:    postRepo,
		AuthClient:  authClient,
	}
}

func (uc *CommentUseCase) CreateComment(ctx context.Context, comment *entity.Comment) error {

	_, err := uc.PostRepo.GetPostByID(ctx, comment.PostID)
	if err != nil {
		return err
	}

	userResp, err := uc.AuthClient.GetUserProfile(ctx, &pb.GetUserProfileRequest{UserId: comment.AuthorID})
	if err != nil || userResp == nil || userResp.User == nil {
		return errors.New("failed to get user info")
	}

	comment.AuthorName = userResp.User.Username
	return uc.CommentRepo.CreateComment(ctx, comment)
}

func (uc *CommentUseCase) GetCommentsByPostID(ctx context.Context, postID int64) ([]entity.Comment, error) {

	_, err := uc.PostRepo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	comments, err := uc.CommentRepo.GetCommentsByPostID(ctx, postID)
	if err != nil {
		return nil, err
	}

	// Fetch author names for each comment
	for i := range comments {
		userResp, err := uc.AuthClient.GetUserProfile(ctx, &pb.GetUserProfileRequest{UserId: comments[i].AuthorID})
		if err != nil || userResp == nil || userResp.User == nil {
			comments[i].AuthorName = "Unknown"
			continue
		}
		comments[i].AuthorName = userResp.User.Username
	}

	return comments, nil
}

func (uc *CommentUseCase) DeleteComment(ctx context.Context, id int64, userID int64) error {
	comment, err := uc.CommentRepo.GetCommentByID(ctx, id)
	if err != nil {
		return err
	}

	if comment.AuthorID != userID {
		return ErrForbidden
	}

	return uc.CommentRepo.DeleteComment(ctx, id)
}

func (uc *CommentUseCase) GetAuthClient() pb.AuthServiceClient {
	return uc.AuthClient
}

func (uc *CommentUseCase) GetComment(ctx context.Context, id int64) (*entity.Comment, error) {
	return uc.CommentRepo.GetCommentByID(ctx, id)
}