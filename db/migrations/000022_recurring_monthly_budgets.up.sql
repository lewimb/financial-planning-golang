-- MONTHLY budgets become recurring (no fixed month/year — they apply every
-- month until deleted). YEARLY budgets keep a required year (still one cap
-- per calendar year) but never had a month.
-- NOTE: DROP NOT NULL must run before the UPDATE that nulls year, or the
-- UPDATE violates the still-active NOT NULL constraint.
ALTER TABLE budgets ALTER COLUMN year DROP NOT NULL;

UPDATE budgets SET month = NULL, year = NULL WHERE period = 'MONTHLY';

-- Existing data may already contain duplicate budgets that the new partial
-- unique indexes below would reject (e.g. two MONTHLY rows for the same
-- user+category, created before this constraint existed). Soft-delete all
-- but the most recently created row in each group.
WITH ranked_monthly AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY user_id, category
               ORDER BY created_at DESC, id DESC
           ) AS rn
    FROM budgets
    WHERE period = 'MONTHLY' AND deleted_at IS NULL
)
UPDATE budgets SET deleted_at = NOW()
WHERE id IN (SELECT id FROM ranked_monthly WHERE rn > 1);

WITH ranked_yearly AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY user_id, category, year
               ORDER BY created_at DESC, id DESC
           ) AS rn
    FROM budgets
    WHERE period = 'YEARLY' AND deleted_at IS NULL
)
UPDATE budgets SET deleted_at = NOW()
WHERE id IN (SELECT id FROM ranked_yearly WHERE rn > 1);

ALTER TABLE budgets DROP CONSTRAINT budgets_user_id_category_period_month_year_key;

ALTER TABLE budgets ADD CONSTRAINT budgets_period_scope_check CHECK (
    (period = 'MONTHLY' AND month IS NULL AND year IS NULL)
    OR (period = 'YEARLY' AND month IS NULL AND year IS NOT NULL)
);

-- One recurring MONTHLY budget per category; one YEARLY budget per
-- category per year. Partial indexes because plain UNIQUE treats every
-- NULL as distinct, which would let duplicates through.
CREATE UNIQUE INDEX budgets_monthly_uniq ON budgets (user_id, category)
    WHERE period = 'MONTHLY' AND deleted_at IS NULL;

CREATE UNIQUE INDEX budgets_yearly_uniq ON budgets (user_id, category, year)
    WHERE period = 'YEARLY' AND deleted_at IS NULL;
