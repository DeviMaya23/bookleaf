CREATE TABLE folder_shares (
    id         UUID PRIMARY KEY,
    folder_id  UUID NOT NULL UNIQUE REFERENCES folders(id) ON DELETE CASCADE,
    token      TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
