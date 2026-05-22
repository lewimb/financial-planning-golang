DROP INDEX IF EXISTS idx_transactions_recurring;
ALTER TABLE transactions
    DROP COLUMN IF EXISTS recurrence_interval,
    DROP COLUMN IF EXISTS is_recurring;
