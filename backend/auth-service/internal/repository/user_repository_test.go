package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jaliks17/ffffforum/backend/auth-service/internal/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	tests := []struct {
		name    string
		user    *entity.User
		mock    func()
		want    int64
		wantErr bool
	}{
		{
			name: "successful creation",
			user: &entity.User{
				Username: "test@example.com",
				Password: "hashed_password",
				Role:     "user",
			},
			mock: func() {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("test@example.com", "hashed_password", "user", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "database error",
			user: &entity.User{
				Username: "test@example.com",
				Password: "hashed_password",
				Role:     "user",
			},
			mock: func() {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("test@example.com", "hashed_password", "user", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := repo.Create(context.Background(), tt.user)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, int64(0), got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	tests := []struct {
		name    string
		id      int64
		mock    func()
		want    *entity.User
		wantErr bool
	}{
		{
			name: "user found",
			id:   1,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "password", "role", "created_at", "updated_at"}).
					AddRow(1, "test@example.com", "hashed_password", "user", time.Now(), time.Now())
				mock.ExpectQuery("SELECT id, username, password, role, created_at, updated_at FROM users").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: &entity.User{
				ID:       1,
				Username: "test@example.com",
				Password: "hashed_password",
				Role:     "user",
			},
			wantErr: false,
		},
		{
			name: "user not found",
			id:   999,
			mock: func() {
				mock.ExpectQuery("SELECT id, username, password, role, created_at, updated_at FROM users").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := repo.GetByID(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				if tt.want != nil {
					assert.Equal(t, tt.want.ID, got.ID)
					assert.Equal(t, tt.want.Username, got.Username)
					assert.Equal(t, tt.want.Password, got.Password)
					assert.Equal(t, tt.want.Role, got.Role)
				} else {
					assert.Nil(t, got)
				}
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	tests := []struct {
		name    string
		email   string
		mock    func()
		want    *entity.User
		wantErr bool
	}{
		{
			name:  "user found",
			email: "test@example.com",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "password", "role", "created_at", "updated_at"}).
					AddRow(1, "test@example.com", "hashed_password", "user", time.Now(), time.Now())
				mock.ExpectQuery("SELECT id, username, password, role, created_at, updated_at FROM users").
					WithArgs("test@example.com").
					WillReturnRows(rows)
			},
			want: &entity.User{
				ID:       1,
				Username: "test@example.com",
				Password: "hashed_password",
				Role:     "user",
			},
			wantErr: false,
		},
		{
			name:  "user not found",
			email: "nonexistent@example.com",
			mock: func() {
				mock.ExpectQuery("SELECT id, username, password, role, created_at, updated_at FROM users").
					WithArgs("nonexistent@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := repo.GetByEmail(context.Background(), tt.email)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				if tt.want != nil {
					assert.Equal(t, tt.want.ID, got.ID)
					assert.Equal(t, tt.want.Username, got.Username)
					assert.Equal(t, tt.want.Password, got.Password)
					assert.Equal(t, tt.want.Role, got.Role)
				} else {
					assert.Nil(t, got)
				}
			}
		})
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	tests := []struct {
		name    string
		username string
		mock    func()
		want    *entity.User
		wantErr bool
	}{
		{
			name:  "user found",
			username: "testuser",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "password", "role", "created_at", "updated_at"}).
					AddRow(1, "testuser", "hashed_password", "user", time.Now(), time.Now())
				mock.ExpectQuery("SELECT id, username, password, role, created_at, updated_at FROM users WHERE username = \\$1").
					WithArgs("testuser").
					WillReturnRows(rows)
			},
			want: &entity.User{
				ID:       1,
				Username: "testuser",
				Password: "hashed_password",
				Role:     "user",
			},
			wantErr: false,
		},
		{
			name:  "user not found",
			username: "nonexistentuser",
			mock: func() {
				mock.ExpectQuery("SELECT id, username, password, role, created_at, updated_at FROM users WHERE username = \\$1").
					WithArgs("nonexistentuser").
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:  "database error",
			username: "testuser",
			mock: func() {
				mock.ExpectQuery("SELECT id, username, password, role, created_at, updated_at FROM users WHERE username = \\$1").
					WithArgs("testuser").
					WillReturnError(assert.AnError)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := repo.GetByUsername(context.Background(), tt.username)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				if tt.want != nil {
					assert.Equal(t, tt.want.ID, got.ID)
					assert.Equal(t, tt.want.Username, got.Username)
					assert.Equal(t, tt.want.Password, got.Password)
					assert.Equal(t, tt.want.Role, got.Role)
				} else {
					assert.Nil(t, got)
				}
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	tests := []struct {
		name    string
		user    *entity.User
		mock    func()
		wantErr bool
	}{
		{
			name: "successful update",
			user: &entity.User{
				ID:       1,
				Username: "updated@example.com",
				Password: "new_hashed_password",
				Role:     "admin",
			},
			mock: func() {
				mock.ExpectExec("UPDATE users").
					WithArgs("updated@example.com", "new_hashed_password", "admin", sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "user not found",
			user: &entity.User{
				ID:       999,
				Username: "nonexistent@example.com",
				Password: "hashed_password",
				Role:     "user",
			},
			mock: func() {
				mock.ExpectExec("UPDATE users").
					WithArgs("nonexistent@example.com", "hashed_password", "user", sqlmock.AnyArg(), 999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := repo.Update(context.Background(), tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	tests := []struct {
		name    string
		id      int64
		mock    func()
		wantErr bool
	}{
		{
			name: "successful deletion",
			id:   1,
			mock: func() {
				mock.ExpectExec("DELETE FROM users").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "user not found",
			id:   999,
			mock: func() {
				mock.ExpectExec("DELETE FROM users").
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := repo.Delete(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
