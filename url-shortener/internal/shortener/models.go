package shortener

import "time"

type URLMapping struct {
    ID        int
    LongURL   string
    ShortURL  string
    CreatedAt time.Time
}

type URLRepository interface {
    Save(mapping URLMapping) error
    FindByShortURL(ShortURL string) (*URLMapping, error)
}