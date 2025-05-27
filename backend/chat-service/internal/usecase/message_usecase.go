package usecase

import (
	"time"

	"github.com/jaliks17/ffffforum/backend/chat-service/internal/entity"
	"github.com/jaliks17/ffffforum/backend/chat-service/internal/repository"
)

type MessageUseCase interface {
	SaveMessage(msg *entity.Message) error
	GetMessages() ([]entity.Message, error)
	DeleteOldMessages(before time.Time) error
}

type messageUseCase struct {
	repo repository.MessageRepository
}

func NewMessageUseCase(repo repository.MessageRepository) MessageUseCase {
	return &messageUseCase{repo: repo}
}

func (uc *messageUseCase) SaveMessage(msg *entity.Message) error {
	return uc.repo.SaveMessage(msg)
}

func (uc *messageUseCase) GetMessages() ([]entity.Message, error) {
	return uc.repo.GetMessages()
}

func (uc *messageUseCase) DeleteOldMessages(before time.Time) error {
	return uc.repo.DeleteOldMessages(before)
}