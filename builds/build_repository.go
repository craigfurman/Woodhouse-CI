package builds

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/craigfurman/woodhouse-ci/jobs"
)

type Repository struct {
	BuildsDir string
}

func (r *Repository) Create(jobId string) (int, io.WriteCloser, chan uint32, error) {
	errs := func(err error) (int, io.WriteCloser, chan uint32, error) {
		return -1, nil, nil, err
	}

	if err := os.MkdirAll(filepath.Join(r.BuildsDir, jobId), os.FileMode(0755)); err != nil {
		return errs(fmt.Errorf("creating builds directory for job %s: %v", jobId, err))
	}

	f, err := os.Create(filepath.Join(r.BuildsDir, jobId, "1-output.txt"))
	if err != nil {
		return errs(fmt.Errorf("creating output file: %v", err))
	}

	status := make(chan uint32, 1)
	go r.recordStatus(jobId, status)

	return 1, f, status, nil
}

func (r *Repository) recordStatus(jobId string, c <-chan uint32) {
	exitCode := <-c
	f, err := os.Create(filepath.Join(r.BuildsDir, jobId, "1-status.txt"))
	if err != nil {
		log.Printf("error creating status file: %v", err)
	}
	_, err = f.Write([]byte(fmt.Sprintf("%d", exitCode)))
	if err != nil {
		log.Printf("error writing status file: %v", err)
	}
	if err := f.Close(); err != nil {
		log.Printf("error closing status file: %v", err)
	}
}

func (r *Repository) Find(jobId string, buildNumber int) (jobs.Build, error) {
	if _, err := os.Stat(filepath.Join(r.BuildsDir, jobId)); os.IsNotExist(err) {
		return jobs.Build{}, fmt.Errorf("no builds found for job %s", jobId)
	}

	out, err := ioutil.ReadFile(filepath.Join(r.BuildsDir, jobId, "1-output.txt"))
	if err != nil {
		return jobs.Build{}, fmt.Errorf("reading output file for job %s. Cause: %v", jobId, err)
	}

	finished, exitStatus, err := r.getBuildExitStatus(jobId)
	if err != nil {
		return jobs.Build{}, err
	}

	return jobs.Build{
		Output:     out,
		ExitStatus: exitStatus,
		Finished:   finished,
	}, nil
}

func (r *Repository) getBuildExitStatus(jobId string) (bool, uint32, error) {
	statusFile := filepath.Join(r.BuildsDir, jobId, "1-status.txt")
	if _, err := os.Stat(statusFile); os.IsNotExist(err) {
		return false, 0, nil
	}

	statusFileContents, err := ioutil.ReadFile(statusFile)
	if err != nil {
		return false, 0, fmt.Errorf("reading status file for job %s. Cause: %v", jobId, err)
	}
	status, err := strconv.Atoi(string(statusFileContents))
	if err != nil {
		return false, 0, fmt.Errorf("converting exit status to integer: %s. Cause: %v", string(statusFileContents), err)
	}
	return true, uint32(status), nil
}
