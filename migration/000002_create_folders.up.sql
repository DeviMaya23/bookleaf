CREATE TABLE folders (
    id UUID PRIMARY KEY,
    user_id TEXT NOT NULL,
    parent_id UUID,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_folders_user
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT fk_folders_parent
        FOREIGN KEY (parent_id) REFERENCES folders (id) ON DELETE RESTRICT
);

CREATE INDEX idx_folders_user_id ON folders (user_id);
CREATE INDEX idx_folders_parent_id ON folders (parent_id);
CREATE INDEX idx_folders_deleted_at ON folders (deleted_at);
