ALTER TABLE images ADD COLUMN is_uploaded boolean NOT NULL DEFAULT false;
UPDATE images SET is_uploaded = true;
