ALTER TABLE bids
    ADD CONSTRAINT bids_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE lots
    ADD CONSTRAINT lots_winner_id_fkey
    FOREIGN KEY (winner_id) REFERENCES users(id) ON DELETE SET NULL;
