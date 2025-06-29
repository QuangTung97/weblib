CREATE TABLE auth_user
(
    id         INTEGER NOT NULL PRIMARY KEY,
    username   TEXT    NOT NULL,
    created_at INTEGER NOT NULL
) STRICT;