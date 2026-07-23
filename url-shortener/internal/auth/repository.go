package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrEmailTaken      = errors.New("email already registered")
	ErrSessionNotFound = errors.New("session not found")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateUser inserts a new user. Relies on the unique constraint on
// email for atomic collision detection — same pattern as the short
// code insert-first approach used for urls.
func (r *Repository) CreateUser(ctx context.Context, name, email, passwordHash string) (*User, error) {
	const q = `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, password_hash, plan, created_at`

	var u User
	err := r.db.QueryRow(ctx, q, name, email, passwordHash).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Plan, &u.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	const q = `
		SELECT id, name, email, password_hash, plan, created_at
		FROM users WHERE email = $1`

	var u User
	err := r.db.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Plan, &u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (*User, error) {
	const q = `
		SELECT id, name, email, password_hash, plan, created_at
		FROM users WHERE id = $1`

	var u User
	err := r.db.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Plan, &u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CreateSession(ctx context.Context, token string, userID int64, expiresAt time.Time) error {
	const q = `
		INSERT INTO sessions (token, user_id, expires_at)
		VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, q, token, userID, expiresAt)
	return err
}

func (r *Repository) GetSession(ctx context.Context, token string) (*Session, error) {
	const q = `
		SELECT token, user_id, expires_at, created_at
		FROM sessions WHERE token = $1`

	var s Session
	err := r.db.QueryRow(ctx, q, token).Scan(&s.Token, &s.UserID, &s.ExpiresAt, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) DeleteSession(ctx context.Context, token string) error {
	const q = `DELETE FROM sessions WHERE token = $1`
	_, err := r.db.Exec(ctx, q, token)
	return err
}

// DeleteExpiredSessions can be run on a schedule (cron/goroutine
// ticker) to keep the sessions table from growing unbounded.
func (r *Repository) DeleteExpiredSessions(ctx context.Context) error {
	const q = `DELETE FROM sessions WHERE expires_at < now()`
	_, err := r.db.Exec(ctx, q)
	return err
}