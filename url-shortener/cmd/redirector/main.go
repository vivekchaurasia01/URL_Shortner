package main

import (
    "context"
    "log"
    "net/http"
    "os"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/redis/go-redis/v9"
    "github.com/vivekchaurasia01/url-shortener/internal/analytics"
    "github.com/vivekchaurasia01/url-shortener/internal/cache"
    "github.com/vivekchaurasia01/url-shortener/internal/redirector"
)


func main () {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}
	redisAddr := os.Getenv("REDIS_ADDR") 
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	    pool, err := pgxpool.New(context.Background(), dbURL)
    if err != nil {
        log.Fatalf("could not connect to database: %v", err)
    }
    defer pool.Close()

    redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})

    repo := redirector.NewURLRepository(pool)
    redisCache := cache.NewRedisCache(redisAddr)
    consumer := analytics.NewConsumer(pool, redisClient)

    // analytics consumer runs in background — never blocks redirect path
    go consumer.Start(context.Background())

    handler := redirector.NewHandler(repo, redisCache, consumer)

    mux := http.NewServeMux()
    mux.HandleFunc("/", handler.Redirect)

    log.Println("redirector service running on :8081")
    log.Fatal(http.ListenAndServe(":8081", mux))
}