PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS albums (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    cover_photo_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS photos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    album_id INTEGER NOT NULL,
    filename TEXT NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    size_bytes INTEGER NOT NULL,
    format TEXT NOT NULL CHECK(format IN ('webp', 'avif')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (album_id) REFERENCES albums(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS share_links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT UNIQUE NOT NULL,
    target_type TEXT NOT NULL CHECK(target_type IN ('album', 'photo')),
    target_id INTEGER NOT NULL,
    max_views INTEGER,
    expires_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    revoked_at DATETIME
);

CREATE TABLE IF NOT EXISTS share_link_views (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    share_link_id INTEGER NOT NULL,
    viewer_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (share_link_id) REFERENCES share_links(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS activity_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT NOT NULL,
    album_id INTEGER,
    photo_id INTEGER,
    share_link_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_photos_album_id ON photos(album_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_share_links_token ON share_links(token);
CREATE UNIQUE INDEX IF NOT EXISTS idx_share_link_views_dedup ON share_link_views(share_link_id, viewer_hash);
CREATE INDEX IF NOT EXISTS idx_activity_events_created_at ON activity_events(created_at);
