
-- +goose Up
CREATE TABLE jobs(
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	command TEXT NOT NULL
);


-- +goose Down
DROP TABLE jobs;
