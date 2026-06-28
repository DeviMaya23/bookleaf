CREATE TABLE ai_categorisation_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    image_id        UUID REFERENCES images(id) ON DELETE SET NULL,
    user_id         TEXT NOT NULL,
    reasoning       TEXT NOT NULL,
    folder_id       UUID,
    new_folder_name TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
