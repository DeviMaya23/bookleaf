CREATE TABLE tags (
    id         UUID        PRIMARY KEY,
    user_id    TEXT        NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, name)
);

CREATE TABLE image_tags (
    image_id UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    tag_id   UUID NOT NULL REFERENCES tags(id)   ON DELETE CASCADE,
    PRIMARY KEY (image_id, tag_id)
);
