ALTER TABLE images
    ADD COLUMN description TEXT,
    ADD COLUMN width INTEGER,
    ADD COLUMN height INTEGER,
    ADD COLUMN file_size BIGINT;
