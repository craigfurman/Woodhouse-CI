
-- +goose Up
CREATE TABLE jobs(
	id TEXT PRIMARY KEY,
	name TEXT
);


-- +goose Down
DROP TABLE jobs;
