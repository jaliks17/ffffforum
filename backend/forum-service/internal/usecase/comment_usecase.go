package usecase

import (
	"context"
	"forum-service/internal/entity"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *entity.Comment) error
	GetByID(ctx context.Context, id int64) (*entity.Comment, error)
	Delete(ctx context.Context, id int64) error
	GetCommentsByPostID(ctx context.Context, postID int64) ([]entity.Comment, error)
	DeleteWithUserID(ctx context.Context, commentID int64, userID int64) error
}

type PostRepository interface {
	GetPostByID(ctx context.Context, id int64) (*entity.Post, error)
}

type CommentUseCase struct {
	repo     CommentRepository
	postRepo PostRepository
}

func NewCommentUseCase(repo CommentRepository, postRepo PostRepository) *CommentUseCase {
	return &CommentUseCase{
		repo:     repo,
		postRepo: postRepo,
	}
}

func (uc *CommentUseCase) CreateComment(ctx context.Context, comment *entity.Comment) error {
	return uc.repo.Create(ctx, comment)
}

func (uc *CommentUseCase) GetComment(ctx context.Context, id int64) (*entity.Comment, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *CommentUseCase) GetCommentsByPostID(ctx context.Context, postID int64) ([]entity.Comment, error) {
	_, err := uc.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	return uc.repo.GetCommentsByPostID(ctx, postID)
}

func (uc *CommentUseCase) DeleteComment(ctx context.Context, commentID int64, userID int64) error {
	err := uc.repo.DeleteWithUserID(ctx, commentID, userID)
	return err
}