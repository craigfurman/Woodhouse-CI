package jobs

import "fmt"

type Job struct {
	ID      string
	Name    string
	Command string
}

type RunningJob struct {
	Job
	Output     string
	ExitStatus uint32
}

//go:generate counterfeiter -o fake_job_repository/fake_job_repository.go . Repository
type Repository interface {
	Save(job *Job) error
	FindById(id string) (Job, error)
}

//go:generate counterfeiter -o fake_job_runner/fake_job_runner.go . Runner
type Runner interface {
	Run(job Job) (RunningJob, error)
}

type Service struct {
	Repository Repository
	Runner     Runner
}

func (s *Service) Save(job *Job) error {
	return s.Repository.Save(job)
}

func (s *Service) RunJob(id string) (RunningJob, error) {
	job, err := s.Repository.FindById(id)
	if err != nil {
		return RunningJob{}, fmt.Errorf("running job with ID: %s. Cause: %v", id, err)
	}

	rj, err := s.Runner.Run(job)
	if err != nil {
		return RunningJob{}, fmt.Errorf("starting job with ID: %s. Cause: %v", id, err)
	}

	return rj, nil
}
