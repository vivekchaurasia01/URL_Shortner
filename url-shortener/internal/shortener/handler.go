package shortener

import (
    "encoding/json"
    "net/http"
)


type Handler struct {
    repo URLRepository
}

func NewHandler(repo URLRepository) *Handler {
    return &Handler{repo: repo}
}

type shortenRequest struct {
    URL string `json:"url"`
}

type shortenResponse struct {
    ShortURL string `json:"short_url"`
}

type statsResponse struct {
	ShortURL string `json:"short_url"`
	Clicks   int64  `json:"clicks"`
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req shortenRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    if req.URL == "" {
        http.Error(w, "url is required", http.StatusBadRequest)
        return
    }

    shortURL, err := GenerateShortURL(req.URL, h.repo)
    if err != nil {
        http.Error(w, "could not shorten url", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(shortenResponse{ShortURL: shortURL})
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		http.Error(w, "code is required", http.StatusBadRequest)
		return
	}

	count, err := h.repo.GetClickCount(code)
	if err != nil {
		http.Error(w, "could not fetch stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statsResponse{ShortURL: code, Clicks: count})
}