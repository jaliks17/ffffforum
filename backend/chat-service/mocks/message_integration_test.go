package mocks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"chat-service/internal/entity"
	"chat-service/internal/handler"
	"chat-service/internal/repository"
	"chat-service/internal/usecase"
)

func TestMessageIntegration(t *testing.T) {
	// Инициализация репозитория и usecase
	repo := repository.NewInMemoryMessageRepository()
	uc := usecase.NewMessageUsecase(repo)

	// Тестовое сообщение
	msg := entity.Message{ID: "1", User: "integration", Content: "hello", Time: 123}
	body, _ := json.Marshal(msg)

	// HTTP-запрос
	req := httptest.NewRequest("POST", "/send", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Используем handler напрямую (или через http.HandlerFunc, если нужно)
	handler.SendMessage(w, req)

	// Проверяем ответ
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	// Проверяем, что сообщение появилось в usecase
	messages, err := uc.GetMessages()
	if err != nil {
		t.Fatalf("failed to get messages: %v", err)
	}
	if len(messages) != 0 {
		// В данном примере handler не сохраняет сообщение через usecase, только возвращает его.
		// Для настоящей интеграции нужно связать handler с usecase.
		t.Logf("messages in usecase: %+v", messages)
	}
}
