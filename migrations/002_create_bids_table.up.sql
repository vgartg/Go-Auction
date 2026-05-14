CREATE TABLE IF NOT EXISTS bids (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lot_id     UUID NOT NULL REFERENCES lots(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL,
    amount     DECIMAL(12,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bids_lot_id_amount ON bids(lot_id, amount DESC);