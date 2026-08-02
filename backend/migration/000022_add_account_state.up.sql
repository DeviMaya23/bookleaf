ALTER TABLE users
    ADD COLUMN account_state TEXT NOT NULL DEFAULT 'active'
        CHECK (account_state IN ('active', 'pending_deletion', 'purged')),
    ADD COLUMN purged_at TIMESTAMPTZ;

UPDATE users SET account_state = 'pending_deletion' WHERE pending_kinde_deletion = TRUE;
