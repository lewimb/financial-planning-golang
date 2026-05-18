-- transactions, budgets, and goals all support UPDATE operations but have no audit
-- timestamp. Add updated_at so the repos can track when a record last changed.
-- NOTE: existing repository UPDATE statements must also SET updated_at = NOW()
-- after this migration is applied.
ALTER TABLE transactions ADD COLUMN updated_at TIMESTAMP;
ALTER TABLE budgets      ADD COLUMN updated_at TIMESTAMP;
ALTER TABLE goals        ADD COLUMN updated_at TIMESTAMP;
