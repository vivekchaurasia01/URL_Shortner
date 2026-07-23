package redirector

import (
    "errors"
    "net/http"
)

// ClickPusher keeps the handler decoupled from the analytics package......
type ClickPusher interface {
    Push(shortURL string)
}

type Handler struct {
    repo   URLRepository
    cache  CacheRepository
    pusher ClickPusher
}

func NewHandler(repo URLRepository, cache CacheRepository, pusher ClickPusher) *Handler {
    return &Handler{repo: repo, cache: cache, pusher: pusher}
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
    shortURL := r.URL.Path[1:] // strip leading /

    if shortURL == "" {
        http.Error(w, "short url required", http.StatusBadRequest)
        return
    }

    longURL, err := ResolveLongURL(shortURL, h.cache, h.repo)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            http.Error(w, "url not found", http.StatusNotFound)
            return
        }
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }

    // fire click event — happens on BOTH cache hit and cache miss paths
    // non-blocking, never delays the redirect response
    h.pusher.Push(shortURL)

    // 302 not 301 — browser won't cache it, every click hits our server
    // essential for analytics — 301 would silently swallow clicks
    http.Redirect(w, r, longURL, http.StatusFound)
}