CREATE TABLE IF NOT EXISTS scan_history (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    source        TEXT    NOT NULL,
    scanned_at    DATETIME NOT NULL DEFAULT (datetime('now')),
    endpoint_count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS collections (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL,
    source     TEXT    NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    json_path  TEXT
);
