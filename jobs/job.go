package jobs

type Job struct {
	ID            string
	Name          string
	GitRepository string
	DockerImage   string
	Command       string
}

type JobStatus int

const (
	NeverRun JobStatus = iota
	Passing
	Failing
)

type JobSummary struct {
	Job
	Status  JobStatus
	Running bool
}

type Build struct {
	Job
	Finished   bool
	Output     []byte
	ExitStatus uint32
}
