CREATE TABLE IF NOT EXISTS goals (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    target_amount INTEGER NOT NULL,
    current_amount INTEGER DEFAULT 0,
    deadline DATE,
    status VARCHAR(20) DEFAULT 'ONGOING', -- ONGOING / COMPLETED
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id)
); 