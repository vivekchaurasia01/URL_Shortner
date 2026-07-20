package redirector
import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

type URLPostgresRepository struct {
    db *pgxpool.Pool
}

func NewURLRepository(db *pgxpool.Pool) URLRepository {
    return &URLPostgresRepository{db: db}
}

func (r *URLPostgresRepository) FindByShortURL(shortURL string) (*URLMapping, error) {
    row := r.db.QueryRow(
        context.Background(),
        "SELECT id, longurl, shorturl,created_at FROM urls WHERE shorturl = $1",
        shortURL,
    )
    var m URLMapping
    if err := row.Scan(&m.LongURL, &m.ShortURL); err != nil {
        return nil, err
    }
    return &m, nil
} 