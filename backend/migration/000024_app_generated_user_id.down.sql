-- This migration is not safely reversible because the original TEXT id values
-- are preserved in idp_subject, but reversing the UUID PK swap requires
-- reconstructing the original schema from scratch.
-- Reverting this migration requires restoring from a pre-migration backup.
SELECT 1; -- no-op placeholder; do not run this down migration in production
