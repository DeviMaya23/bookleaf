-- River schema migrations 004 (remainder), 005, 006.
-- Runs in a separate transaction so that the 'pending' enum value added
-- by migration 012 is committed and safe to reference here.

-- 004 (remainder): rebuild finalized constraint to include pending,
-- create river_queue and river_leader adjustments
ALTER TABLE river_job DROP CONSTRAINT finalized_or_finalized_at_null;
ALTER TABLE river_job ADD CONSTRAINT finalized_or_finalized_at_null CHECK (
    (finalized_at IS NULL AND state NOT IN ('cancelled', 'completed', 'discarded')) OR
    (finalized_at IS NOT NULL AND state IN ('cancelled', 'completed', 'discarded'))
);

CREATE TABLE river_queue (
    name text PRIMARY KEY NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    metadata jsonb NOT NULL DEFAULT '{}' ::jsonb,
    paused_at timestamptz,
    updated_at timestamptz NOT NULL
);

ALTER TABLE river_leader
    ALTER COLUMN name SET DEFAULT 'default',
    DROP CONSTRAINT name_length,
    ADD CONSTRAINT name_length CHECK (name = 'default');

-- 005: rebuild migration table and add unique_key
DO
$body$
BEGIN
    IF (SELECT to_regclass('river_migration') IS NOT NULL) THEN
        ALTER TABLE river_migration RENAME TO river_migration_old;

        CREATE TABLE river_migration(
            line TEXT NOT NULL,
            version bigint NOT NULL,
            created_at timestamptz NOT NULL DEFAULT NOW(),
            CONSTRAINT line_length CHECK (char_length(line) > 0 AND char_length(line) < 128),
            CONSTRAINT version_gte_1 CHECK (version >= 1),
            PRIMARY KEY (line, version)
        );

        INSERT INTO river_migration (created_at, line, version)
        SELECT created_at, 'main', version
        FROM river_migration_old;

        DROP TABLE river_migration_old;
    END IF;
END;
$body$
LANGUAGE 'plpgsql';

ALTER TABLE river_job ADD COLUMN IF NOT EXISTS unique_key bytea;

CREATE UNIQUE INDEX IF NOT EXISTS river_job_kind_unique_key_idx ON river_job (kind, unique_key) WHERE unique_key IS NOT NULL;

CREATE UNLOGGED TABLE river_client (
    id text PRIMARY KEY NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    metadata jsonb NOT NULL DEFAULT '{}',
    paused_at timestamptz,
    updated_at timestamptz NOT NULL,
    CONSTRAINT name_length CHECK (char_length(id) > 0 AND char_length(id) < 128)
);

CREATE UNLOGGED TABLE river_client_queue (
    river_client_id text NOT NULL REFERENCES river_client (id) ON DELETE CASCADE,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    max_workers bigint NOT NULL DEFAULT 0,
    metadata jsonb NOT NULL DEFAULT '{}',
    num_jobs_completed bigint NOT NULL DEFAULT 0,
    num_jobs_running bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (river_client_id, name),
    CONSTRAINT name_length CHECK (char_length(name) > 0 AND char_length(name) < 128),
    CONSTRAINT num_jobs_completed_zero_or_positive CHECK (num_jobs_completed >= 0),
    CONSTRAINT num_jobs_running_zero_or_positive CHECK (num_jobs_running >= 0)
);

-- 006: bulk unique
CREATE OR REPLACE FUNCTION river_job_state_in_bitmask(bitmask BIT(8), state river_job_state)
RETURNS boolean
LANGUAGE SQL
IMMUTABLE
AS $$
    SELECT CASE state
        WHEN 'available' THEN get_bit(bitmask, 7)
        WHEN 'cancelled' THEN get_bit(bitmask, 6)
        WHEN 'completed' THEN get_bit(bitmask, 5)
        WHEN 'discarded' THEN get_bit(bitmask, 4)
        WHEN 'pending'   THEN get_bit(bitmask, 3)
        WHEN 'retryable' THEN get_bit(bitmask, 2)
        WHEN 'running'   THEN get_bit(bitmask, 1)
        WHEN 'scheduled' THEN get_bit(bitmask, 0)
        ELSE 0
    END = 1;
$$;

ALTER TABLE river_job ADD COLUMN IF NOT EXISTS unique_states BIT(8);

CREATE UNIQUE INDEX IF NOT EXISTS river_job_unique_idx ON river_job (unique_key)
    WHERE unique_key IS NOT NULL
      AND unique_states IS NOT NULL
      AND river_job_state_in_bitmask(unique_states, state);

DROP INDEX river_job_kind_unique_key_idx;
