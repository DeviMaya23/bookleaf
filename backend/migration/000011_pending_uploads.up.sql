DELETE FROM images WHERE is_uploaded = false;
ALTER TABLE images DROP COLUMN is_uploaded;

CREATE TABLE pending_uploads (
    id          UUID        NOT NULL PRIMARY KEY,
    user_id     TEXT        NOT NULL REFERENCES users(id),
    title       TEXT        NOT NULL,
    description TEXT,
    source_url  TEXT,
    r2_path     TEXT        NOT NULL,
    mime_type   TEXT        NOT NULL,
    folder_id   UUID        REFERENCES folders(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_pending_uploads_user_id ON pending_uploads (user_id);
CREATE INDEX idx_pending_uploads_created_at ON pending_uploads (created_at);
