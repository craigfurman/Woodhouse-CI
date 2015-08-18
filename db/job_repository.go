package db

import (
	"database/sql"
	"fmt"

	"github.com/craigfurman/woodhouse-ci/jobs"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pborman/uuid"
)

type JobRepository struct {
	db *sql.DB
}

func NewJobRepository(dbPath string) (*JobRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return &JobRepository{
		db: db,
	}, nil
}

func (repo *JobRepository) Save(job *jobs.Job) error {
	job.ID = uuid.New()
	_, err := repo.db.Exec(
		"INSERT INTO jobs(id, name, command, dockerimage, gitrepository) VALUES(?, ?, ?, ?, ?)",
		job.ID,
		job.Name,
		job.Command,
		job.DockerImage,
		job.GitRepository,
	)
	return err
}

func (repo *JobRepository) FindById(id string) (jobs.Job, error) {
	job := jobs.Job{ID: id}
	if err := repo.db.QueryRow("SELECT name, command, dockerimage, gitrepository FROM jobs WHERE id=?", id).
		Scan(&job.Name, &job.Command, &job.DockerImage, &job.GitRepository); err != nil {
		return jobs.Job{}, fmt.Errorf("no job found with ID: %s. Cause: %v", id, err)
	}
	return job, nil
}

func (repo *JobRepository) Close() error {
	return repo.db.Close()
}
