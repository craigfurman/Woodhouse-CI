package db_test

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/craigfurman/woodhouse-ci/db"
	"github.com/craigfurman/woodhouse-ci/jobs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JobRepository", func() {

	var repo *db.JobRepository

	BeforeEach(func() {
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		dbPath := filepath.Join(cwd, "sqlite", "store.db")
		os.Remove(dbPath)
		migrateCmd := exec.Command("goose", "up")
		migrateCmd.Dir = filepath.Join(cwd, "..")
		migrateCmd.Stdout = GinkgoWriter
		migrateCmd.Stderr = GinkgoWriter
		Expect(migrateCmd.Run()).To(Succeed())

		repo, err = db.NewJobRepository(dbPath)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(repo.Close()).To(Succeed())
	})

	Describe("creating a job", func() {
		var (
			savedJob   *jobs.Job
			saveJobErr error
		)

		BeforeEach(func() {
			savedJob = &jobs.Job{
				Name:          "myFancyJob",
				Command:       "my CI script",
				DockerImage:   "someUser/someName:someTag",
				GitRepository: "sweet potato",
			}
			saveJobErr = repo.Save(savedJob)
		})

		It("does not error", func() {
			Expect(saveJobErr).NotTo(HaveOccurred())
		})

		It("generates a uuid for the job", func() {
			Expect(savedJob.ID).ToNot(BeEmpty())
		})

		Describe("listing jobs", func() {
			It("lists jobs", func() {
				list, err := repo.List()
				Expect(err).NotTo(HaveOccurred())
				Expect(list).To(ConsistOf(jobs.Job{
					ID:            savedJob.ID,
					Name:          "myFancyJob",
					Command:       "my CI script",
					DockerImage:   "someUser/someName:someTag",
					GitRepository: "sweet potato",
				}))
			})

			// TODO how to simulate error here?
			PContext("when listing jobs fails", func() {})
		})

		Describe("retrieving the job", func() {
			It("retrieves the job", func() {
				job, err := repo.FindById(savedJob.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(job).To(Equal(jobs.Job{
					ID:            savedJob.ID,
					Name:          "myFancyJob",
					Command:       "my CI script",
					DockerImage:   "someUser/someName:someTag",
					GitRepository: "sweet potato",
				}))
			})

			Context("when no job with that ID exists", func() {
				It("returns error", func() {
					_, err := repo.FindById("idontexist")
					Expect(err).To(MatchError(ContainSubstring("no job found with ID: idontexist")))
				})
			})
		})
	})
})
