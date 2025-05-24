package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"auth-service/internal/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewSessionRepository(sqlxDB)

	tests := []struct {
		name    string
		session *entity.Session
		mock    func()
		want    int64
		wantErr bool
	}{
		{
			name: "successful creation",
			session: &entity.Session{
				UserID:    1,
				Token:     "test-token",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			mock: func() {
				mock.ExpectQuery("INSERT INTO sessions").
					WithArgs(1, "test-token", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "database error",
			session: &entity.Session{
				UserID:    1,
				Token:     "test-token",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			mock: func() {
				mock.ExpectQuery("INSERT INTO sessions").
					WithArgs(1, "test-token", sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := repo.Create(context.Background(), tt.session)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, int64(0), tt.session.ID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, tt.session.ID)
			}
		})
	}
}

func TestSessionRepository_GetByToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewSessionRepository(sqlxDB)

	tests := []struct {
		name    string
		token   string
		mock    func()
		want    *entity.Session
		wantErr bool
	}{
		{
			name:  "session found",
			token: "test-token",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id", "token", "expires_at"}).
					AddRow(1, 1, "test-token", time.Now().Add(time.Hour))
				mock.ExpectQuery("SELECT id, user_id, token, expires_at FROM sessions").
					WithArgs("test-token").
					WillReturnRows(rows)
			},
			want: &entity.Session{
				ID:        1,
				UserID:    1,
				Token:     "test-token",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			wantErr: false,
		},
		{
			name:  "session not found",
			token: "nonexistent-token",
			mock: func() {
				mock.ExpectQuery("SELECT id, user_id, token, expires_at FROM sessions").
					WithArgs("nonexistent-token").
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := repo.GetByToken(context.Background(), tt.token)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				if tt.want != nil {
					assert.Equal(t, tt.want.ID, got.ID)
					assert.Equal(t, tt.want.UserID, got.UserID)
					assert.Equal(t, tt.want.Token, got.Token)
				} else {
					assert.Nil(t, got)
				}
			}
		})
	}
}

func TestSessionRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewSessionRepository(sqlxDB)

	tests := []struct {
		name    string
		token   string
		mock    func()
		wantErr bool
	}{
		{
			name:  "successful deletion",
			token: "test-token",
			mock: func() {
				mock.ExpectExec("DELETE FROM sessions").
					WithArgs("test-token").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:  "session not found",
			token: "nonexistent-token",
			mock: func() {
				mock.ExpectExec("DELETE FROM sessions").
					WithArgs("nonexistent-token").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := repo.Delete(context.Background(), tt.token)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSessionRepository_DeleteExpired(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewSessionRepository(sqlxDB)

	tests := []struct {
		name    string
		mock    func()
		wantErr bool
	}{
		{
			name: "successful deletion of expired sessions",
			mock: func() {
				mock.ExpectExec("DELETE FROM sessions").
					WithArgs(sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(5, 5))
			},
			wantErr: false,
		},
		{
			name: "no expired sessions",
			mock: func() {
				mock.ExpectExec("DELETE FROM sessions").
					WithArgs(sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := repo.DeleteExpired(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
