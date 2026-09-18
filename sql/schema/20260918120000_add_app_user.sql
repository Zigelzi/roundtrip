-- +goose Up
CREATE TABLE app_user (
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
INSERT INTO app_user (id, name) VALUES (1, 'Parent 1');

-- +goose Down
DROP TABLE app_user;
