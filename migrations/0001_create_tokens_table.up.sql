CREATE TABLE tokens (
                      id SERIAL PRIMARY KEY,
                      session TEXT NOT NULL,
                      refresh_hash TEXT NOT NULL
);

CREATE INDEX tokens_session_hash_idx ON tokens USING HASH(session);