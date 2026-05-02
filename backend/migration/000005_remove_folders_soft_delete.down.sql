ALTER TABLE images
    DROP CONSTRAINT fk_images_folder;

ALTER TABLE images
    ADD CONSTRAINT fk_images_folder
        FOREIGN KEY (folder_id) REFERENCES folders (id) ON DELETE SET NULL;

ALTER TABLE folders
    ADD COLUMN deleted_at TIMESTAMPTZ;

CREATE INDEX idx_folders_deleted_at ON folders (deleted_at);
