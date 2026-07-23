package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

const sessionCookieName = "session_token"

type Handler struct {
	service *Service
	// secureCookies should be true in production (HTTPS) and false
	// for local http://localhost dev — browsers drop Secure cookies
	// on plain HTTP.
	secureCookies bool
}

func NewHandler(service *Service, secureCookies bool) *Handler {
	return &Handler{service: service, secureCookies: secureCookies}
}

// Routes registers all auth endpoints on the given mux. Call this
// from cmd/shortener/main.go: auth.NewHandler(svc, prod).Routes(mux)
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("POST /api/auth/logout", h.Logout)
	mux.HandleFunc("GET /api/me", h.Me)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, token, expiresAt, err := h.service.Register(r.Context(), req)
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	h.setSessionCookie(w, token, expiresAt)
	writeJSON(w, http.StatusCreated, user.ToResponse())
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, token, expiresAt, err := h.service.Login(r.Context(), req)
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	h.setSessionCookie(w, token, expiresAt)
	writeJSON(w, http.StatusOK, user.ToResponse())
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		if err := h.service.Logout(r.Context(), cookie.Value); err != nil {
			slog.Warn("logout: failed to delete session", "error", err)
		}
	}
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "not logged in")
		return
	}

	user, err := h.service.CurrentUser(r.Context(), cookie.Value)
	if err != nil {
		h.clearSessionCookie(w)
		writeError(w, http.StatusUnauthorized, "session expired")
		return
	}

	writeJSON(w, http.StatusOK, user.ToResponse())
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,                 // JS on the frontend never reads this
		Secure:   h.secureCookies,      // true in prod (HTTPS only)
		SameSite: http.SameSiteLaxMode, // sent on top-level navigation (OAuth redirects), blocked on cross-site POSTs
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrEmailTaken):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrInvalidEmail), errors.Is(err, ErrPasswordTooShort), errors.Is(err, ErrNameRequired):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		slog.Error("auth handler error", "error", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
	}
}

// ---- small JSON helpers, local to this package to avoid coupling
// it to whatever response helpers cmd/shortener already has ----

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message, "message": message})
}