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
	_, err := repo.db.Exec("INSERT INTO jobs(id, name) VALUES(?, ?)", job.ID, job.Name)
	return err
}

func (repo *JobRepository) FindById(id string) (jobs.Job, error) {
	var name string
	if err := repo.db.QueryRow("SELECT name FROM jobs WHERE id=?", id).Scan(&name); err != nil {
		return jobs.Job{}, fmt.Errorf("no job found with ID: %s. Cause: %v", id, err)
	}
	return jobs.Job{ID: id, Name: name}, nil
}

func (repo *JobRepository) Close() error {
	return repo.db.Close()
}
