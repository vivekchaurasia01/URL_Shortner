package redirector

import "errors"

var ErrNotFound = errors.New("short URL not found")  //New returns an error that formats as the given text. 

type URLRepository interface {
    FindByShortURL(shortURL string) (*URLMapping, error)
}

type CacheRepository interface {
    Get(shortURL string) (string, error)
    Set(shortURL string, longURL string) error
}

func ResolveLongURL(shortURL string, cache CacheRepository, repo URLRepository) (string, error) {
    // check Redis first....
    longURL, err := cache.Get(shortURL)
    if err == nil {
        // short url found in redis, DB never touched....
        return longURL, nil
    }

    // cache miss — request forward to Postgres.......
    mapping, err := repo.FindByShortURL(shortURL)
    if err != nil {
        return "", ErrNotFound
    }

    // fill cache so next request for this code is a hit...
    cache.Set(shortURL, mapping.LongURL)

    return mapping.LongURL, nil
}