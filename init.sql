CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    original VARCHAR(2048) NOT NULL,
    clicks INT DEFAULT 0
);