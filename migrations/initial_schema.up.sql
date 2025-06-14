BEGIN TRANSACTION;

CREATE TYPE STATUS AS ENUM ('open','in progress','done');
CREATE TYPE USER_ROLE AS ENUM ('writer','reader','admin');

CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY,
    email VARCHAR(150) NOT NULL UNIQUE,
    role USER_ROLE NOT NULL
);

CREATE TABLE IF NOT EXISTS lists(
    id UUID PRIMARY KEY,
    name VARCHAR(250) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL,
    last_updated TIMESTAMP NOT NULL,
    owner UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS todos(
    id UUID PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    description VARCHAR(250) NOT NULL,
    list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    status STATUS NOT NULL,
    created_at TIMESTAMP NOT NULL,
    last_updated TIMESTAMP NOT NULL,
    assigned_to UUID REFERENCES users(id),
    due_date TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_lists(
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    list_id UUID REFERENCES lists(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id,list_id)
);

CREATE INDEX idx_todos_list_id ON todos(list_id);
CREATE INDEX idx_lists_owner_id ON lists(owner);

COMMIT;
