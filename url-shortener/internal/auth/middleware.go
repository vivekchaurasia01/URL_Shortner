package auth

import (
	"context"
	"net/http"
)

type contextKey string

const userContextKey contextKey = "auth_user"

// WithOptionalUser attaches the logged-in user to the request context
// if a valid session cookie is present, but lets the request through
// either way — use this on /api/shorten, which anonymous users can
// hit too, but which needs to know the caller's plan for gating
// custom aliases / QR / expiry rules.
func (h *Handler) WithOptionalUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		user, err := h.service.CurrentUser(r.Context(), cookie.Value)
		if err != nil {
			next.ServeHTTP(w, r) // invalid/expired session — treat as anonymous, don't block
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireUser rejects the request with 401 if there's no valid
// session — use this on endpoints that must have a logged-in user,
// e.g. /api/billing/subscribe.
func (h *Handler) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "login required")
			return
		}
		user, err := h.service.CurrentUser(r.Context(), cookie.Value)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "login required")
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserFromContext reads the user attached by WithOptionalUser or
// RequireUser. Returns nil if the caller is anonymous.
func UserFromContext(ctx context.Context) *User {
	user, _ := ctx.Value(userContextKey).(*User)
	return user
}