package helpers

import "github.com/craigfurman/woodhouse-ci/jobs"

func JobGrid(list []jobs.Job) [][]jobs.Job {
	grid := [][]jobs.Job{}
	row := []jobs.Job{}
	for _, job := range list {
		if len(row) == 3 {
			grid = append(grid, row)
			row = []jobs.Job{}
		}

		row = append(row, job)
	}
	if len(row) > 0 {
		grid = append(grid, row)
	}

	return grid
}
