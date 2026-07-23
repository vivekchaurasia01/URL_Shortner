package auth

import "time"

// User is the internal representation — never serialize this directly,
// use UserResponse so PasswordHash never leaves the service layer.
type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
	Plan         string // "free" | "pro"
	CreatedAt    time.Time
}

// Session is a DB-backed opaque session, keyed by a random token
// (not a JWT — simplest option that fits the current stack; can be
// swapped for signed tokens later without touching the frontend).
type Session struct {
	Token     string
	UserID    int64
	ExpiresAt time.Time
	CreatedAt time.Time
}

const SessionTTL = 7 * 24 * time.Hour

// ---- request DTOs (decoded from JSON bodies) ----

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ---- response DTOs (what the frontend actually sees) ----

type UserResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Plan  string `json:"plan"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		Plan:  u.Plan,
	}
}