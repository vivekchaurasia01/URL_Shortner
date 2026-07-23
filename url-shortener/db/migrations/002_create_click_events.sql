CREATE TABLE IF NOT EXISTS click_events (
    id          BIGSERIAL PRIMARY KEY,
    shorturl    VARCHAR(7) NOT NULL,
    clicked_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_click_events_shorturl  ON click_events (shorturl);
CREATE INDEX idx_click_events_clicked_at ON click_events (clicked_at);