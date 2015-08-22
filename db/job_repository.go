package db

import (
	"database/sql"
	"fmt"

	"github.com/craigfurman/woodhouse-ci/jobs"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pborman/uuid"
)

type JobRepository struct {
	db          *sql.DB
	SkeletonKey []byte
}

func NewJobRepository(dbPath, skeletonKey string) (*JobRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return &JobRepository{
		db:          db,
		SkeletonKey: hash(skeletonKey),
	}, nil
}

func (repo *JobRepository) Save(job *jobs.Job) error {
	job.ID = uuid.New()

	encryptedGitPassword, encryptedGitPasswordIv, err := repo.encrypt(job.GitPassword)
	if err != nil {
		fmt.Errorf("error encrypting git password for job %s, cause: %v\n", job.ID, err)
	}
	_, err = repo.db.Exec(
		"INSERT INTO jobs(id, name, command, dockerimage, gitrepository, gitpassword, gitpasswordiv) VALUES(?, ?, ?, ?, ?, ?, ?)",
		job.ID,
		job.Name,
		job.Command,
		job.DockerImage,
		job.GitRepository,
		encryptedGitPassword,
		encryptedGitPasswordIv,
	)
	return err
}

func (repo *JobRepository) FindById(id string) (jobs.Job, error) {
	job := jobs.Job{ID: id}

	var encryptedGitPassword, encryptedGitPasswordIv string

	if err := repo.db.QueryRow("SELECT name, command, dockerimage, gitrepository, gitpassword, gitpasswordiv FROM jobs WHERE id=?", id).
		Scan(&job.Name, &job.Command, &job.DockerImage, &job.GitRepository, &encryptedGitPassword, &encryptedGitPasswordIv); err != nil {
		return jobs.Job{}, fmt.Errorf("no job found with ID: %s. Cause: %v", id, err)
	}

	decryptedGitPassword, err := repo.decrypt(encryptedGitPassword, encryptedGitPasswordIv)
	if err != nil {
		return jobs.Job{}, fmt.Errorf("error decrypting git password for job %s, cause: %v\n", job.ID, err)
	}

	job.GitPassword = decryptedGitPassword
	return job, nil
}

func (repo *JobRepository) Close() error {
	return repo.db.Close()
}
