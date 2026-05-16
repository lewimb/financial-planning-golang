CREATE TABLE IF NOT EXISTS budgets (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    category VARCHAR(255) NOT NULL,
    period VARCHAR(10) NOT NULL CHECK (period IN ('MONTHLY', 'YEARLY')),
    month INTEGER,
    year INTEGER NOT NULL,
    limit_amount INTEGER NOT NULL,
    alert_threshold INTEGER DEFAULT 80,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id),

    -- prevent duplicate budgets
    UNIQUE (user_id, category, period, month, year)
);