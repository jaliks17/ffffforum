package repository

import (
	"chat-service/internal/entity"
	"testing"
)

func TestInMemoryMessageRepository(t *testing.T) {
	repo := NewInMemoryMessageRepository()
	msg := entity.Message{ID: "1", User: "test", Content: "hello", Time: 123}
	if err := repo.Save(msg); err != nil {
		t.Fatalf("failed to save message: %v", err)
	}
	messages, err := repo.GetAll()
	if err != nil {
		t.Fatalf("failed to get messages: %v", err)
	}
	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}
}
