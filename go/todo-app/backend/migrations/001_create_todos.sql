-- Creates the todos table used by the Go backend.
CREATE TABLE IF NOT EXISTS todos (
    id          SERIAL PRIMARY KEY,
    title       TEXT NOT NULL,
    completed   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index on completed for efficient filtering (optional but useful).
CREATE INDEX IF NOT EXISTS idx_todos_completed ON todos(completed);
