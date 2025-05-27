package repository

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/jaliks17/ffffforum/backend/chat-service/internal/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSaveMessage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewMessageRepository(db)

	tests := []struct {
		name    string
		msg     entity.Message
		mock    func()
		wantErr bool
	}{
		{
			name: "successful message save",
			msg: entity.Message{
				UserID: 1,
				Username: "testuser",
				Message:  "Hello world",
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO chat_messages (user_id, username, content, timestamp) VALUES ($1, $2, $3, $4) RETURNING id")).
					WithArgs(1, "testuser", "Hello world", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "database error on save",
			msg: entity.Message{
				UserID: 2,
				Username: "testuser",
				Message:  "Hello world",
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO chat_messages (user_id, username, content, timestamp) VALUES ($1, $2, $3, $4) RETURNING id")).
					WithArgs(2, "testuser", "Hello world", sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "empty author name",
			msg: entity.Message{
				UserID: 3,
				Username: "",
				Message:  "test",
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO chat_messages (user_id, username, content, timestamp) VALUES ($1, $2, $3, $4) RETURNING id")).
					WithArgs(3, "", "test", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := repo.SaveMessage(&tt.msg)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetMessages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewMessageRepository(db)

	expectedQuery := regexp.QuoteMeta("SELECT id, user_id, username, content as message, timestamp FROM chat_messages ORDER BY timestamp ASC")

	tests := []struct {
		name    string
		mock    func()
		want    []entity.Message
		wantErr bool
	}{
		{
			name: "successful get messages",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id", "username", "message", "timestamp"}).
					AddRow(1, 101, "user1", "message 1", time.Now()).
					AddRow(2, 102, "user2", "message 2", time.Now())
				mock.ExpectQuery(expectedQuery).
					WillReturnRows(rows)
			},
			want: []entity.Message{
				{ID: 1, UserID: 101, Username: "user1", Message: "message 1"},
				{ID: 2, UserID: 102, Username: "user2", Message: "message 2"},
			},
			wantErr: false,
		},
		{
			name: "empty result",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id", "username", "message", "timestamp"})
				mock.ExpectQuery(expectedQuery).
					WillReturnRows(rows)
			},
			want:    []entity.Message{},
			wantErr: false,
		},
		{
			name: "scan error",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id", "username"}).
					AddRow(1, 101, "user1")
				mock.ExpectQuery(expectedQuery).
					WillReturnRows(rows)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			messages, err := repo.GetMessages()
			assert.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, messages) // Expect nil messages on error
			} else {
				assert.NoError(t, err)
				if len(tt.want) == 0 {
					// If expecting an empty result, accept either nil or an empty slice
					if messages != nil {
						assert.Len(t, messages, 0)
					} else {
						assert.Nil(t, messages)
					}
				} else {
					// If expecting non-empty result, compare content
					assert.NotNil(t, messages) // Expect non-nil messages if expecting content
					actualMessagesWithoutTimestamp := make([]entity.Message, len(messages))
					for i, msg := range messages {
						actualMessagesWithoutTimestamp[i] = entity.Message{
							ID:       msg.ID,
							UserID:    msg.UserID,
							Username: msg.Username,
							Message:  msg.Message,
						}
					}
					assert.Equal(t, tt.want, actualMessagesWithoutTimestamp)
				}
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMessageRepository_DeleteOldMessages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewMessageRepository(db)

	cutoffTime := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name    string
		mock    func()
		wantErr bool
	}{
		{
			name: "successful deletion",
			mock: func() {
				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM chat_messages WHERE timestamp < $1")).
					WithArgs(sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "database error",
			mock: func() {
				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM chat_messages WHERE timestamp < $1")).
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := repo.DeleteOldMessages(cutoffTime)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}