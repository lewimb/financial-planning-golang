CREATE TABLE user_financial_profiles (
    id                SERIAL PRIMARY KEY,
    user_id           INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    monthly_income    NUMERIC(15, 2) NOT NULL DEFAULT 0,
    fixed_expenses    NUMERIC(15, 2) NOT NULL DEFAULT 0,
    current_savings   NUMERIC(15, 2) NOT NULL DEFAULT 0,
    debt              NUMERIC(15, 2) NOT NULL DEFAULT 0,
    employment_status VARCHAR(100)   NOT NULL,
    spending_habit    VARCHAR(100),
    risk_level        VARCHAR(50),
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE TABLE user_financial_goals (
    id        SERIAL PRIMARY KEY,
    user_id   INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    goal_type VARCHAR(100) NOT NULL,
    UNIQUE (user_id, goal_type)
);

CREATE INDEX idx_user_financial_profiles_user_id ON user_financial_profiles (user_id);
CREATE INDEX idx_user_financial_goals_user_id    ON user_financial_goals (user_id);
