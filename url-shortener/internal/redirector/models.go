package redirector

import "time"


type URLMapping struct {
    ID        int
    LongURL   string
    ShortURL  string
    CreatedAt time.Time
}

