PRAGMA foreign_keys = true;

CREATE TABLE migrations (
    key TEXT NOT NULL PRIMARY KEY,
    value INTEGER
);

CREATE TABLE users (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    created_at INTEGER NOT NULL DEFAULT unixepoch(),
    username TEXT NOT NULL,
    -- Temporary code for the "state" param on Tesla's auth flow
    registration_code TEXT,
    -- "sub" field from Tesla's ID token
    tesla_id TEXT
);

CREATE TABLE vehicles (
    vin TEXT NOT NULL PRIMARY KEY,
    created_at INTEGER NOT NULL DEFAULT unixepoch(),
    owner INTEGER REFERENCES users(id)
);

INSERT INTO migrations (key, value) VALUES ('version', 1);
