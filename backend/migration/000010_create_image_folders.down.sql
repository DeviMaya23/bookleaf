ALTER TABLE images ADD COLUMN folder_id UUID REFERENCES folders(id) ON DELETE SET NULL;

UPDATE images
SET folder_id = sub.folder_id
FROM (
    SELECT DISTINCT ON (image_id) image_id, folder_id
    FROM image_folders
    ORDER BY image_id, position ASC
) sub
WHERE images.id = sub.image_id;

DROP TABLE image_folders;
