
-- +goose Up
CREATE TABLE jobs(
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	command TEXT NOT NULL,
	dockerimage TEXT NOT NULL,
	gitrepository TEXT NOT NULL,
	gitpassword TEXT NOT NULL,
	gitpasswordiv TEXT NOT NULL
);


-- +goose Down
DROP TABLE jobs;
