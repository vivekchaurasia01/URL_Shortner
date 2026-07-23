package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vivekchaurasia01/url-shortener/internal/auth"
	"github.com/vivekchaurasia01/url-shortener/internal/shortener"
)

const (
	listenAddr        = ":8080"
	frontendOrigin    = "http://127.0.0.1:5500" // Live Server dev origin — change for prod
	secureCookies     = false                    // flip to true once served over HTTPS
)

// cors allows the frontend (served on a different port during dev)
// to call this API with credentials (the session cookie). The
// allowed origin must be an exact match, not "*", because credentials
// are involved — browsers reject wildcard + credentials together.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", frontendOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	dbStr := os.Getenv("DATABASE_URL")
	if dbStr == "" {
		log.Fatal("DATABASE_URL not set")
	}

	pool, err := pgxpool.New(context.Background(), dbStr)
	if err != nil {
		log.Fatalf("can't connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("can't ping database: %v", err)
	}

	// Repositories
	urlRepo := shortener.NewURLRepository(pool)
	authRepo := auth.NewRepository(pool)

	// Services
	authService := auth.NewService(authRepo)

	// Handlers
	shortenerHandler := shortener.NewHandler(urlRepo)
	authHandler := auth.NewHandler(authService, secureCookies)

	// Router
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/shorten", shortenerHandler.ShortenURL)
	authHandler.Routes(mux)

	srv := &http.Server{
		Addr:         listenAddr,
		Handler:      cors(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("shortener service running on %s", listenAddr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}