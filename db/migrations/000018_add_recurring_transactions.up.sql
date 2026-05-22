ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS is_recurring         BOOLEAN     NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS recurrence_interval  VARCHAR(20);

CREATE INDEX idx_transactions_recurring
    ON transactions(user_id, is_recurring)
    WHERE is_recurring = TRUE AND deleted_at IS NULL;
