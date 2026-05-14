CREATE TABLE IF NOT EXISTS lots (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title         TEXT NOT NULL,
    start_price   DECIMAL(12,2) NOT NULL,
    min_step      DECIMAL(12,2) NOT NULL,
    current_price DECIMAL(12,2) NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    closing_at    TIMESTAMP NOT NULL,
    version       INT NOT NULL DEFAULT 1,
    winner_id     UUID
);

CREATE INDEX idx_lots_status_closing_at ON lots(status, closing_at);