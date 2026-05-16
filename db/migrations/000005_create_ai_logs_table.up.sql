CREATE TABLE IF NOT EXISTS ai_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    question TEXT,
    response TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id)
);