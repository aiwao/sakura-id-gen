CREATE TABLE IF NOT EXISTS accounts (
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    instaddr_id TEXT NOT NULL,
    instaddr_password TEXT NOT NULL
);
