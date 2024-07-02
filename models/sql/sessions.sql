CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    user_id INT UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT UNIQUE NOT NULL
);


INSERT INTO sessions (user_id, token_hash)
VALUES (1, 'xyz-123') ON CONFLICT (user_id) DO
UPDATE
SET token_hash='xyz-123';
