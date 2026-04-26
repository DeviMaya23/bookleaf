CREATE TABLE images (
    id UUID PRIMARY KEY,
    user_id TEXT NOT NULL,
    folder_id UUID,
    title TEXT NOT NULL,
    source_url TEXT,
    r2_path TEXT NOT NULL,
    thumbnail_path TEXT,
    mime_type TEXT NOT NULL,
    ai_labels JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_images_user
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT fk_images_folder
        FOREIGN KEY (folder_id) REFERENCES folders (id) ON DELETE SET NULL
);

CREATE INDEX idx_images_user_id ON images (user_id);
CREATE INDEX idx_images_folder_id ON images (folder_id);
CREATE INDEX idx_images_deleted_at ON images (deleted_at);
