package helpers

import "github.com/craigfurman/woodhouse-ci/jobs"

func JobGrid(list []jobs.Build) [][]jobs.Build {
	grid := [][]jobs.Build{}
	row := []jobs.Build{}
	for _, build := range list {
		if len(row) == 3 {
			grid = append(grid, row)
			row = []jobs.Build{}
		}

		row = append(row, build)
	}
	if len(row) > 0 {
		grid = append(grid, row)
	}

	return grid
}
