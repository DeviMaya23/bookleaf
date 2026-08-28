BEGIN;

-- Step 1: Add idp_subject column to users and populate it from the existing TEXT id
ALTER TABLE users ADD COLUMN idp_subject TEXT;
UPDATE users SET idp_subject = id;
ALTER TABLE users ALTER COLUMN idp_subject SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_idp_subject_unique UNIQUE (idp_subject);

-- Step 2: Add new UUID id columns to all tables with user_id TEXT FKs
ALTER TABLE users ADD COLUMN new_id UUID DEFAULT gen_random_uuid();
ALTER TABLE folders ADD COLUMN new_user_id UUID;
ALTER TABLE images ADD COLUMN new_user_id UUID;
ALTER TABLE tags ADD COLUMN new_user_id UUID;
ALTER TABLE pending_uploads ADD COLUMN new_user_id UUID;
ALTER TABLE ai_categorisation_logs ADD COLUMN new_user_id UUID;

-- Step 3: Backfill FK tables by joining against users.new_id
UPDATE folders f SET new_user_id = (SELECT new_id FROM users u WHERE u.id = f.user_id);
UPDATE images i SET new_user_id = (SELECT new_id FROM users u WHERE u.id = i.user_id);
UPDATE tags t SET new_user_id = (SELECT new_id FROM users u WHERE u.id = t.user_id);
UPDATE pending_uploads p SET new_user_id = (SELECT new_id FROM users u WHERE u.id = p.user_id);
UPDATE ai_categorisation_logs a SET new_user_id = (SELECT new_id FROM users u WHERE u.id = a.user_id);

-- Step 4: Drop existing FK constraints referencing users.id
ALTER TABLE folders DROP CONSTRAINT fk_folders_user;
ALTER TABLE images DROP CONSTRAINT fk_images_user;
ALTER TABLE tags DROP CONSTRAINT tags_user_id_fkey;
ALTER TABLE pending_uploads DROP CONSTRAINT pending_uploads_user_id_fkey;

-- Step 5: Swap the PK on users
ALTER TABLE users DROP CONSTRAINT users_pkey;
ALTER TABLE users DROP COLUMN id;
ALTER TABLE users RENAME COLUMN new_id TO id;
ALTER TABLE users ALTER COLUMN id SET NOT NULL;
ALTER TABLE users ADD PRIMARY KEY (id);

-- Step 6: Swap user_id columns on dependent tables
ALTER TABLE folders DROP COLUMN user_id;
ALTER TABLE folders RENAME COLUMN new_user_id TO user_id;
ALTER TABLE folders ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE folders ADD CONSTRAINT fk_folders_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT;

ALTER TABLE images DROP COLUMN user_id;
ALTER TABLE images RENAME COLUMN new_user_id TO user_id;
ALTER TABLE images ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE images ADD CONSTRAINT fk_images_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT;

ALTER TABLE tags DROP COLUMN user_id;
ALTER TABLE tags RENAME COLUMN new_user_id TO user_id;
ALTER TABLE tags ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE tags ADD CONSTRAINT tags_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE pending_uploads DROP COLUMN user_id;
ALTER TABLE pending_uploads RENAME COLUMN new_user_id TO user_id;
ALTER TABLE pending_uploads ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE pending_uploads ADD CONSTRAINT pending_uploads_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id);

ALTER TABLE ai_categorisation_logs DROP COLUMN user_id;
ALTER TABLE ai_categorisation_logs RENAME COLUMN new_user_id TO user_id;

COMMIT;
