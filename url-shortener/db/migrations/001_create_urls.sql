CREATE TABLE IF NOT EXISTS urls (
    id          BIGSERIAL PRIMARY KEY,
    longurl     TEXT NOT NULL,
    shorturl    VARCHAR(7) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT urls_shorturl_unique UNIQUE (shorturl)
);




