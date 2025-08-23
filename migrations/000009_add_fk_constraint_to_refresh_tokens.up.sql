ALTER TABLE user_refresh_tokens
ADD CONSTRAINT user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;