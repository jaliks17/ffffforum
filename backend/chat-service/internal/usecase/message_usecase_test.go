package usecase

import (
	"errors"
	"testing"

	"chat-service/internal/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) SaveMessage(msg *entity.Message) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockMessageRepository) GetMessages() ([]entity.Message, error) {
	args := m.Called()
	return args.Get(0).([]entity.Message), args.Error(1)
}

func TestMessageUseCase_SaveMessage(t *testing.T) {
	mockRepo := new(MockMessageRepository)
	uc := NewMessageUseCase(mockRepo)

	msg1 := &entity.Message{UserID: 1, Username: "test", Message: "hello"}
	mockRepo.On("SaveMessage", msg1).Return(nil)
	err := uc.SaveMessage(msg1)
	assert.NoError(t, err)

	msg2 := &entity.Message{UserID: 2, Username: "error", Message: "fail"}
	mockRepo.On("SaveMessage", msg2).Return(errors.New("db error"))
	err = uc.SaveMessage(msg2)
	assert.Error(t, err)
}

func TestMessageUseCase_GetMessages(t *testing.T) {
	mockRepo := new(MockMessageRepository)
	uc := NewMessageUseCase(mockRepo)

	expected := []entity.Message{{ID: 1, UserID: 1, Username: "user", Message: "test"}}
	mockRepo.On("GetMessages").Return(expected, nil)
	result, err := uc.GetMessages()
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	mockRepo.On("GetMessages").Return([]entity.Message{}, nil)
	result, err = uc.GetMessages()
	assert.NoError(t, err)
	assert.Empty(t, result)
}
