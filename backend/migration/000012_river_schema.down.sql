-- Reverse River schema migrations 004 (partial), 003, 002, 001.
-- Note: the 'pending' enum value cannot be removed from river_job_state
-- once added; rolling back leaves it in place (harmless without the schema
-- that uses it).

-- 004 (partial) down
ALTER TABLE river_job ALTER COLUMN args DROP NOT NULL;
ALTER TABLE river_job ALTER COLUMN metadata DROP NOT NULL;
ALTER TABLE river_job ALTER COLUMN metadata DROP DEFAULT;

-- 003 down
ALTER TABLE river_job
    ALTER COLUMN tags DROP NOT NULL,
    ALTER COLUMN tags DROP DEFAULT;

-- 002 down
DROP TABLE river_job;
DROP FUNCTION IF EXISTS river_job_notify;
DROP TYPE river_job_state;
DROP TABLE river_leader;

-- 001 down
DROP TABLE river_migration;
