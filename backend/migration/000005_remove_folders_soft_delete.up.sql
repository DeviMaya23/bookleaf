DROP INDEX idx_folders_deleted_at;
ALTER TABLE folders
    DROP COLUMN deleted_at;

ALTER TABLE images
    DROP CONSTRAINT fk_images_folder;

ALTER TABLE images
    ADD CONSTRAINT fk_images_folder
        FOREIGN KEY (folder_id) REFERENCES folders (id) ON DELETE RESTRICT;
