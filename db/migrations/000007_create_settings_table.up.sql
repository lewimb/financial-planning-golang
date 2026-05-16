CREATE TABLE IF NOT EXISTS settings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE,
    currency VARCHAR(10) DEFAULT 'IDR',
    language VARCHAR(10) DEFAULT 'EN',
    notification_enabled BOOLEAN DEFAULT TRUE,

    FOREIGN KEY (user_id) REFERENCES users(id)
);