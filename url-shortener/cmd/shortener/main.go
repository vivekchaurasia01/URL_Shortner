package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vivekchaurasia01/url-shortener/internal/shortener"
)


func main () {
	dbStr := os.Getenv("DATABASE_URL")  // we didnt write our string directly because we dont want it to leak on git after push..

	if dbStr == "" {
		log.Fatal("DATABASE_URL not set")
	}

	pool,err := pgxpool.New(context.Background(),dbStr)
	if err != nil {
		log.Fatalf("cant connect to database : %v",err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("cant ping database : %v",err)
	}

	repo := shortener.NewURLRepository(pool)
	handler := shortener.NewHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("/shorten",handler.ShortenURL)

	log.Println("shortener service running on :8080")
	log.Fatal(http.ListenAndServe(":8080",mux))


}