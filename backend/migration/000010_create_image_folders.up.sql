CREATE TABLE image_folders (
    image_id  UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    folder_id UUID NOT NULL REFERENCES folders(id) ON DELETE CASCADE,
    position  TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (image_id, folder_id)
);

CREATE INDEX idx_image_folders_folder_id ON image_folders (folder_id);
CREATE INDEX idx_image_folders_image_id  ON image_folders (image_id);
CREATE INDEX idx_image_folders_folder_position ON image_folders (folder_id, position);

INSERT INTO image_folders (image_id, folder_id, position)
SELECT id, folder_id,
    ROW_NUMBER() OVER (PARTITION BY folder_id ORDER BY created_at ASC)::TEXT
FROM images
WHERE folder_id IS NOT NULL;

ALTER TABLE images DROP COLUMN folder_id;
