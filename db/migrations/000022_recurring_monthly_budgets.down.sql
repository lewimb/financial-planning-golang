DROP INDEX IF EXISTS budgets_yearly_uniq;
DROP INDEX IF EXISTS budgets_monthly_uniq;

ALTER TABLE budgets DROP CONSTRAINT IF EXISTS budgets_period_scope_check;

-- Best-effort backfill: the specific month/year a MONTHLY budget used to
-- have is not recoverable, so pin rows back to the current calendar month
-- just to satisfy the old NOT NULL/UNIQUE shape.
UPDATE budgets SET year = EXTRACT(YEAR FROM CURRENT_DATE)::int WHERE year IS NULL;
UPDATE budgets SET month = EXTRACT(MONTH FROM CURRENT_DATE)::int WHERE period = 'MONTHLY' AND month IS NULL;

-- The old unique constraint (unlike the partial indexes it's about to
-- replace) doesn't ignore soft-deleted rows, so a soft-deleted duplicate
-- backfilled to the same month/year as its active sibling would violate
-- it. Soft-deleted budgets have zero live behavior anywhere in the app
-- (every query already filters deleted_at IS NULL) — hard-delete them
-- here rather than inventing fake distinguishing month/year values just
-- to satisfy a constraint being restored on a rollback path.
DELETE FROM budgets WHERE deleted_at IS NOT NULL;

ALTER TABLE budgets ALTER COLUMN year SET NOT NULL;

ALTER TABLE budgets ADD CONSTRAINT budgets_user_id_category_period_month_year_key
    UNIQUE (user_id, category, period, month, year);
