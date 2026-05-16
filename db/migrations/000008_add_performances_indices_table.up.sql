-- transactions
CREATE INDEX idx_transactions_user_category_date
ON transactions(user_id, category, date);

-- optional but good
CREATE INDEX idx_transactions_full
ON transactions(user_id, category, type, date);

-- budgets
CREATE INDEX idx_budgets_user
ON budgets(user_id);