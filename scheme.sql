CREATE TABLE IF NOT EXISTS accounts (
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%S%fZ', 'now'))
);
