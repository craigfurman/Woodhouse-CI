package jobs

import (
	"fmt"
	"io"

	"github.com/craigfurman/woodhouse-ci/chunkedio"
)

type Job struct {
	ID            string
	Name          string
	GitRepository string
	DockerImage   string
	Command       string
}

type Build struct {
	Job
	Finished   bool
	Output     []byte
	ExitStatus uint32
}

//go:generate counterfeiter -o fake_job_repository/fake_job_repository.go . JobRepository
type JobRepository interface {
	List() ([]Job, error)
	Save(job *Job) error
	FindById(id string) (Job, error)
}

//go:generate counterfeiter -o fake_build_repository/fake_build_repository.go . BuildRepository
type BuildRepository interface {
	Create(jobId string) (int, io.WriteCloser, chan uint32, error)
	Find(jobId string, buildNumber int) (Build, error)
	Stream(jobId string, buildNumber int, startAtByte int64) (*chunkedio.ChunkedReader, error)
}

//go:generate counterfeiter -o fake_job_runner/fake_job_runner.go . Runner
type Runner interface {
	Run(job Job, outputDest io.WriteCloser, status chan<- uint32) error
}

type Service struct {
	JobRepository   JobRepository
	Runner          Runner
	BuildRepository BuildRepository
}

func (s *Service) ListJobs() ([]Job, error) {
	return s.JobRepository.List()
}

func (s *Service) Save(job *Job) error {
	return s.JobRepository.Save(job)
}

func (s *Service) RunJob(id string) error {
	job, err := s.JobRepository.FindById(id)
	if err != nil {
		return fmt.Errorf("running job with ID: %s. Cause: %v", id, err)
	}

	_, outputDest, exitStatusChan, err := s.BuildRepository.Create(id)
	if err != nil {
		return fmt.Errorf("creating build data for job with ID: %s. Cause: %v", id, err)
	}

	if err := s.Runner.Run(job, outputDest, exitStatusChan); err != nil {
		return fmt.Errorf("starting job with ID: %s. Cause: %v", id, err)
	}

	return nil
}

func (s *Service) FindBuild(jobId string, buildNumber int) (Build, error) {
	job, err := s.JobRepository.FindById(jobId)
	if err != nil {
		return Build{}, err
	}
	build, err := s.BuildRepository.Find(jobId, buildNumber)
	if err != nil {
		return Build{}, err
	}
	build.Job = job
	return build, nil
}

func (s *Service) Stream(jobId string, buildNumber int, streamOffset int64) (*chunkedio.ChunkedReader, error) {
	return s.BuildRepository.Stream(jobId, buildNumber, streamOffset)
}
