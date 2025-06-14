BEGIN TRANSACTION;
ALTER TABLE lists CLUSTER ON lists_pkey;
ALTER TABLE todos CLUSTER ON todos_pkey;
ALTER TABLE users CLUSTER ON users_pkey;
ALTER TABLE user_lists CLUSTER ON user_lists_pkey;

CLUSTER lists;
CLUSTER todos;
CLUSTER users;
CLUSTER user_lists;
END;