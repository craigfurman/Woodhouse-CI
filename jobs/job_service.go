package jobs

import (
	"fmt"
	"io"

	"github.com/craigfurman/woodhouse-ci/chunkedio"
)

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
	HighestBuild(jobId string) (int, error)
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

func (s *Service) AllLatestBuilds() ([]Build, error) {
	errs := func(err error) ([]Build, error) {
		return []Build{}, fmt.Errorf("listing all latest builds. cause: %v\n", err)
	}

	jobList, err := s.JobRepository.List()
	if err != nil {
		return errs(err)
	}

	builds := []Build{}
	for _, job := range jobList {
		highestBuildForJob, err := s.HighestBuild(job.ID)
		if err != nil {
			return errs(err)
		}

		build, err := s.findBuild(job, highestBuildForJob)
		if err != nil {
			return errs(err)
		}

		builds = append(builds, build)
	}
	return builds, nil
}

func (s *Service) Save(job *Job) error {
	return s.JobRepository.Save(job)
}

func (s *Service) RunJob(id string) (int, error) {
	job, err := s.JobRepository.FindById(id)
	if err != nil {
		return 0, fmt.Errorf("running job with ID: %s. Cause: %v", id, err)
	}

	buildNumber, outputDest, exitStatusChan, err := s.BuildRepository.Create(id)
	if err != nil {
		return 0, fmt.Errorf("creating build data for job with ID: %s. Cause: %v", id, err)
	}

	if err := s.Runner.Run(job, outputDest, exitStatusChan); err != nil {
		return 0, fmt.Errorf("starting job with ID: %s. Cause: %v", id, err)
	}

	return buildNumber, nil
}

func (s *Service) FindBuild(jobId string, buildNumber int) (Build, error) {
	job, err := s.JobRepository.FindById(jobId)
	if err != nil {
		return Build{}, err
	}
	return s.findBuild(job, buildNumber)
}

func (s *Service) findBuild(job Job, buildNumber int) (Build, error) {
	build, err := s.BuildRepository.Find(job.ID, buildNumber)
	if err != nil {
		return Build{}, err
	}
	build.Job = job
	return build, nil
}

func (s *Service) HighestBuild(jobId string) (int, error) {
	return s.BuildRepository.HighestBuild(jobId)
}

func (s *Service) Stream(jobId string, buildNumber int, streamOffset int64) (*chunkedio.ChunkedReader, error) {
	return s.BuildRepository.Stream(jobId, buildNumber, streamOffset)
}
