package shortener

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
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
	if err != nil {  // err type -> same short url exist || or db down etc.....
		// we need to identify type of error so we can identify if we  should retry or not...
		// check if it's specifically a collision...
		var pgErr *pgconn.PgError
		if errors.As(err,&pgErr) && pgErr.Code == "23505" {  //23505 collison err..
			return ErrDuplicateShortURL
		}
		return err  // other err than collision...
	}
	return nil

	
}

func (r *URLPostregesRepository) FindByShortURL (shortURL string) (*URLMapping, error) {
	
}