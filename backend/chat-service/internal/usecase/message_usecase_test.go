package usecase

import (
	"chat-service/internal/entity"
	"chat-service/internal/repository"
	"testing"
)

func TestMessageUsecase(t *testing.T) {
	repo := repository.NewInMemoryMessageRepository()
	usecase := NewMessageUsecase(repo)
	msg := entity.Message{ID: "1", User: "test", Content: "hello", Time: 123}
	if err := usecase.SendMessage(msg); err != nil {
		t.Fatalf("failed to send message: %v", err)
	}
	messages, err := usecase.GetMessages()
	if err != nil {
		t.Fatalf("failed to get messages: %v", err)
	}
	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}
}
