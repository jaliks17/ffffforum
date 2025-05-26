package repository

import (
	"errors"
	"regexp"
	"testing"

	"chat-service/internal/entity"

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
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO chat_messages (user_id, username, message) VALUES ($1, $2, $3)")).
					WithArgs(1, "testuser", "Hello world").
					WillReturnResult(sqlmock.NewResult(1, 1))
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
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO chat_messages (user_id, username, message) VALUES ($1, $2, $3)")).
					WithArgs(2, "testuser", "Hello world").
					WillReturnError(errors.New("database error"))
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
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO chat_messages (user_id, username, message) VALUES ($1, $2, $3)")).
					WithArgs(3, "", "test").
					WillReturnResult(sqlmock.NewResult(1, 1))
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

	tests := []struct {
		name    string
		mock    func()
		want    []entity.Message
		wantErr bool
	}{
		{
			name: "successful get messages",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id", "username", "message"}).
					AddRow(1, 101, "user1", "message 1").
					AddRow(2, 102, "user2", "message 2")
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, username, message FROM chat_messages")).
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
				rows := sqlmock.NewRows([]string{"id", "user_id", "username", "message"})
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, username, message FROM chat_messages")).
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
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, username, message FROM chat_messages")).
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
			assert.Equal(t, tt.want, messages)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}