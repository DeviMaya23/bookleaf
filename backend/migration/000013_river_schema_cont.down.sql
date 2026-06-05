-- Reverse River schema migrations 006, 005, 004 (remainder).

-- 006 down
DROP INDEX river_job_unique_idx;
ALTER TABLE river_job DROP COLUMN unique_states;
CREATE UNIQUE INDEX IF NOT EXISTS river_job_kind_unique_key_idx ON river_job (kind, unique_key) WHERE unique_key IS NOT NULL;
DROP FUNCTION river_job_state_in_bitmask;

-- 005 down
DO
$body$
BEGIN
    IF (SELECT to_regclass('river_migration') IS NOT NULL) THEN
        IF EXISTS (SELECT * FROM river_migration WHERE line <> 'main') THEN
            RAISE EXCEPTION 'Found non-main migration lines; down migration cannot proceed.';
        END IF;

        ALTER TABLE river_migration RENAME TO river_migration_old;

        CREATE TABLE river_migration(
            id bigserial PRIMARY KEY,
            created_at timestamptz NOT NULL DEFAULT NOW(),
            version bigint NOT NULL,
            CONSTRAINT version CHECK (version >= 1)
        );

        CREATE UNIQUE INDEX ON river_migration USING btree(version);

        INSERT INTO river_migration (created_at, version)
        SELECT created_at, version FROM river_migration_old;

        DROP TABLE river_migration_old;
    END IF;
END;
$body$
LANGUAGE 'plpgsql';

ALTER TABLE river_job DROP COLUMN unique_key;
DROP TABLE river_client_queue;
DROP TABLE river_client;

-- 004 (remainder) down
ALTER TABLE river_job DROP CONSTRAINT finalized_or_finalized_at_null;
ALTER TABLE river_job ADD CONSTRAINT finalized_or_finalized_at_null CHECK (
  (state IN ('cancelled', 'completed', 'discarded') AND finalized_at IS NOT NULL) OR finalized_at IS NULL
);

CREATE OR REPLACE FUNCTION river_job_notify()
  RETURNS TRIGGER
  AS $$
DECLARE
  payload json;
BEGIN
  IF NEW.state = 'available' THEN
    payload = json_build_object('queue', NEW.queue);
    PERFORM pg_notify('river_insert', payload::text);
  END IF;
  RETURN NULL;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER river_notify
  AFTER INSERT ON river_job
  FOR EACH ROW
  EXECUTE PROCEDURE river_job_notify();

DROP TABLE river_queue;

ALTER TABLE river_leader
    ALTER COLUMN name DROP DEFAULT,
    DROP CONSTRAINT name_length,
    ADD CONSTRAINT name_length CHECK (char_length(name) > 0 AND char_length(name) < 128);
