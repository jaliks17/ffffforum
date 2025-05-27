package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	domain "github.com/jaliks17/ffffforum/backend/auth-service/internal/entity"

	"github.com/jmoiron/sqlx"
)

type ISessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByToken(ctx context.Context, token string) (*domain.Session, error)
	Delete(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
}

type SessionRepository struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	query := `
		INSERT INTO sessions (user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		session.UserID,
		session.Token,
		session.ExpiresAt,
		time.Now(),
	).Scan(&session.ID)

	return err
}

func (r *SessionRepository) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM sessions
		WHERE token = $1
	`

	var session domain.Session
	err := r.db.GetContext(ctx, &session, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &session, nil
}

func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	query := `
		DELETE FROM sessions
		WHERE token = $1
	`

	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM sessions
		WHERE expires_at < CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query)
	return err
}