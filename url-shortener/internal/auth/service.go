package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrNameRequired       = errors.New("name is required")
	ErrInvalidCredentials = errors.New("incorrect email or password")
	ErrSessionExpired     = errors.New("session expired")
)

const bcryptCost = 12 // ~250ms/hash on typical hardware — deliberate, not accidental

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Register validates input, hashes the password, creates the user,
// and starts a session for them (so registering logs you straight in,
// matching what register.html expects).
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*User, string, time.Time, error) {
	name := strings.TrimSpace(req.Name)
	email := strings.ToLower(strings.TrimSpace(req.Email))

	if name == "" {
		return nil, "", time.Time{}, ErrNameRequired
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, "", time.Time{}, ErrInvalidEmail
	}
	if len(req.Password) < 8 {
		return nil, "", time.Time{}, ErrPasswordTooShort
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	user, err := s.repo.CreateUser(ctx, name, email, string(hash))
	if err != nil {
		return nil, "", time.Time{}, err // may be ErrEmailTaken
	}

	token, expiresAt, err := s.startSession(ctx, user.ID)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	return user, token, expiresAt, nil
}

// Login verifies credentials and starts a new session.
// Deliberately returns the same ErrInvalidCredentials whether the
// email doesn't exist or the password is wrong, so the API doesn't
// leak which emails are registered.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*User, string, time.Time, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, "", time.Time{}, ErrInvalidCredentials
		}
		return nil, "", time.Time{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, "", time.Time{}, ErrInvalidCredentials
	}

	token, expiresAt, err := s.startSession(ctx, user.ID)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	return user, token, expiresAt, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	return s.repo.DeleteSession(ctx, token)
}

// CurrentUser resolves a session cookie value to a user, or
// ErrSessionExpired/ErrSessionNotFound if it's missing or stale.
func (s *Service) CurrentUser(ctx context.Context, token string) (*User, error) {
	if token == "" {
		return nil, ErrSessionNotFound
	}
	session, err := s.repo.GetSession(ctx, token)
	if err != nil {
		return nil, err
	}
	if session.ExpiresAt.Before(time.Now()) {
		_ = s.repo.DeleteSession(ctx, token) // best-effort cleanup
		return nil, ErrSessionExpired
	}
	return s.repo.GetUserByID(ctx, session.UserID)
}

func (s *Service) startSession(ctx context.Context, userID int64) (string, time.Time, error) {
	token, err := generateToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(SessionTTL)
	if err := s.repo.CreateSession(ctx, token, userID, expiresAt); err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

// generateToken makes a 256-bit random, URL-safe opaque session
// token. This is not a JWT — nothing is encoded in it, it's just an
// unguessable key into the sessions table.
func generateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}