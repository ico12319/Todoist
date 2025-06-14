BEGIN;

CREATE TABLE IF NOT EXISTS user_refresh_tokens(
  refresh_token VARCHAR(255) UNIQUE ,
  user_id UUID REFERENCES users(id) PRIMARY KEY
);

CREATE INDEX idx_refresh_token ON user_refresh_tokens(refresh_token);

COMMIT;