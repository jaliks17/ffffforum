package usecase

import (
	"errors"
	"testing"
	"time"

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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Message), args.Error(1)
}

func (m *MockMessageRepository) DeleteOldMessages(before time.Time) error {
	args := m.Called(before)
	return args.Error(0)
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

	tests := []struct {
		name    string
		mockSetup func(mockRepo *MockMessageRepository)
		want    []entity.Message
		wantErr bool
	}{
		{
			name: "successful get messages",
			mockSetup: func(mockRepo *MockMessageRepository) {
				expected := []entity.Message{{ID: 1, UserID: 1, Username: "user", Message: "test"}}
				mockRepo.On("GetMessages").Return(expected, nil)
			},
			want:    []entity.Message{{ID: 1, UserID: 1, Username: "user", Message: "test"}},
			wantErr: false,
		},
		{
			name: "empty result",
			mockSetup: func(mockRepo *MockMessageRepository) {
				mockRepo.On("GetMessages").Return([]entity.Message{}, nil)
			},
			want:    []entity.Message{},
			wantErr: false,
		},
		// Добавьте другие тест-кейсы, например, для ошибки репозитория
		{
			name: "repository error",
			mockSetup: func(mockRepo *MockMessageRepository) {
				mockRepo.On("GetMessages").Return(nil, errors.New("db error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMessageRepository) // Создаем новый мок для каждого подтеста
			uc := NewMessageUseCase(mockRepo)

			tt.mockSetup(mockRepo)

			result, err := uc.GetMessages()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMessageUseCase_DeleteOldMessages(t *testing.T) {
	tests := []struct {
		name    string
		mockSetup func(mockRepo *MockMessageRepository)
		before  time.Time
		wantErr bool
	}{
		{
			name: "successful deletion",
			mockSetup: func(mockRepo *MockMessageRepository) {
				mockRepo.On("DeleteOldMessages", mock.AnythingOfType("time.Time")).Return(nil).Once()
			},
			before:  time.Now().Add(-24 * time.Hour),
			wantErr: false,
		},
		{
			name: "repository error",
			mockSetup: func(mockRepo *MockMessageRepository) {
				mockRepo.On("DeleteOldMessages", mock.AnythingOfType("time.Time")).Return(errors.New("db error")).Once()
			},
			before:  time.Now().Add(-24 * time.Hour),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMessageRepository)
			uc := NewMessageUseCase(mockRepo)

			tt.mockSetup(mockRepo)

			err := uc.DeleteOldMessages(tt.before)

			assert.Equal(t, tt.wantErr, err != nil)
			mockRepo.AssertExpectations(t)
		})
	}
}
