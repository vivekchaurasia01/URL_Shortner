package shortener

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type URLPostregesRepository struct {
	db *pgxpool.Pool
}

func NewURLRepository(db *pgxpool.Pool) URLRepository {
    return &URLPostregesRepository{db: db}
}

func (r *URLPostregesRepository) Save (mapping URLMapping) error {
	_,err := r.db.Exec (
		context.Background(),
		"INSERT INTO urls (longurl, shorturl) VALUES ($1, $2)",
		mapping.LongURL,
		mapping.ShortURL,
	)

	
}