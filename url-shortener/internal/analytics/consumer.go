package analytics

import (
    "context"
    "log"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/redis/go-redis/v9"
)

type ClickEvent struct {
    ShortURL  string
    Timestamp time.Time
}

type Consumer struct {
    db     *pgxpool.Pool
    redis  *redis.Client
    events chan ClickEvent
}

func NewConsumer(db *pgxpool.Pool, redisClient *redis.Client) *Consumer {
    return &Consumer{
        db:     db,
        redis:  redisClient,
        events: make(chan ClickEvent, 10000), // buffered — redirect never waits on this
    }
}

// Push sends a click event into the async pipeline
// select with default: if channel is full, drop the event rather than blocking redirect
func (c *Consumer) Push(shortURL string) {
    select {
    case c.events <- ClickEvent{ShortURL: shortURL, Timestamp: time.Now()}:
    default:
        log.Printf("analytics channel full, dropping event for %s", shortURL)
    }
}

// Start drains the event channel — run in a separate goroutine
func (c *Consumer) Start(ctx context.Context) {
    for {
        select {
        case event := <-c.events:
            c.process(ctx, event)
        case <-ctx.Done():
            return
        }
    }
}

func (c *Consumer) process(ctx context.Context, event ClickEvent) {
    // append-only click_events table — source of truth for real analytics
    // GROUP BY date, trends, breakdowns all come from querying this table later
    _, err := c.db.Exec(ctx,
        "INSERT INTO click_events (shorturl, clicked_at) VALUES ($1, $2)",
        event.ShortURL,
        event.Timestamp,
    )
    if err != nil {
        log.Printf("failed to insert click event: %v", err)
    }

    // fast running counter in Redis — answers "how many clicks" instantly
    // separate from the analytics table, each does a different job
    if err := c.redis.Incr(ctx, "count:"+event.ShortURL).Err(); err != nil {
        log.Printf("failed to increment redis counter: %v", err)
    }
}