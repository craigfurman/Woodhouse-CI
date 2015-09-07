package helpers_test

import (
	"strconv"

	"github.com/craigfurman/woodhouse-ci/jobs"
	"github.com/craigfurman/woodhouse-ci/web/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Job grid structure", func() {
	var (
		list []jobs.Job
		grid [][]jobs.Job
	)

	job := func(i int) jobs.Job {
		return jobs.Job{ID: strconv.Itoa(i)}
	}

	createList := func(n int) []jobs.Job {
		l := []jobs.Job{}
		for i := 0; i < n; i++ {
			l = append(l, job(i))
		}
		return l
	}

	JustBeforeEach(func() {
		grid = helpers.JobGrid(list)
	})

	Context("when there are no jobs", func() {
		BeforeEach(func() {
			list = createList(0)
		})

		It("returns an empty grid", func() {
			Expect(grid).To(BeEmpty())
		})
	})

	Context("when there is one job", func() {
		BeforeEach(func() {
			list = createList(1)
		})

		It("returns one row with one column", func() {
			Expect(grid).To(Equal(
				[][]jobs.Job{
					{job(0)},
				},
			))
		})
	})

	Context("when there are 4 jobs", func() {
		BeforeEach(func() {
			list = createList(4)
		})

		It("returns 1 row of 3 and 1 row of 1", func() {
			Expect(grid).To(Equal(
				[][]jobs.Job{
					{job(0), job(1), job(2)},
					{job(3)},
				},
			))
		})
	})
})
