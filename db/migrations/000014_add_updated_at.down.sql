ALTER TABLE transactions DROP COLUMN IF EXISTS updated_at;
ALTER TABLE budgets      DROP COLUMN IF EXISTS updated_at;
ALTER TABLE goals        DROP COLUMN IF EXISTS updated_at;
