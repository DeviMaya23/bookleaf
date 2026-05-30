DROP TABLE pending_uploads;
ALTER TABLE images ADD COLUMN is_uploaded BOOLEAN NOT NULL DEFAULT true;
