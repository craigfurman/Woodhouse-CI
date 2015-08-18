package builds

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/craigfurman/woodhouse-ci/blockingio"
)

func (r *Repository) Stream(jobId string, buildNumber int, startAtByte int64) (*blockingio.BlockingReader, error) {
	outputFile, err := os.Open(filepath.Join(r.BuildsDir, jobId, fmt.Sprintf("%d-output.txt", buildNumber)))
	if err != nil {
		return nil, fmt.Errorf("streaming output from job: %s, build: %d. Cause: %v", jobId, buildNumber, err)
	}

	if _, err := outputFile.Seek(startAtByte, 0); err != nil {
		return nil, fmt.Errorf("seeking: cause: %v", err)
	}

	doneWriting := func() bool {
		_, err := os.Stat(filepath.Join(r.BuildsDir, jobId, "1-status.txt"))
		return !os.IsNotExist(err)
	}

	return &blockingio.BlockingReader{Output: outputFile, DoneWriting: doneWriting}, nil
}
