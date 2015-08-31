package jobs_test

import (
	"errors"
	"io"
	"io/ioutil"

	"github.com/craigfurman/woodhouse-ci/jobs"
	"github.com/craigfurman/woodhouse-ci/jobs/fake_build_repository"
	"github.com/craigfurman/woodhouse-ci/jobs/fake_job_repository"
	"github.com/craigfurman/woodhouse-ci/jobs/fake_job_runner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service", func() {
	var (
		service   *jobs.Service
		jobRepo   *fake_job_repository.FakeJobRepository
		buildRepo *fake_build_repository.FakeBuildRepository
		runner    *fake_job_runner.FakeRunner
	)

	BeforeEach(func() {
		jobRepo = new(fake_job_repository.FakeJobRepository)
		buildRepo = new(fake_build_repository.FakeBuildRepository)
		runner = new(fake_job_runner.FakeRunner)
		service = &jobs.Service{
			JobRepository:   jobRepo,
			Runner:          runner,
			BuildRepository: buildRepo,
		}
	})

	Describe("saving a job", func() {
		It("saves the job using the jobRepository", func() {
			Expect(service.Save(&jobs.Job{Name: "freddo", Command: "whoami"})).To(Succeed())
			Expect(jobRepo.SaveCallCount()).To(Equal(1))
			Expect(jobRepo.SaveArgsForCall(0).Name).To(Equal("freddo"))
			Expect(jobRepo.SaveArgsForCall(0).Command).To(Equal("whoami"))
		})

		Context("when saving fails", func() {
			BeforeEach(func() {
				jobRepo.SaveReturns(errors.New("something went wrong"))
			})

			It("returns the error from the jobRepository", func() {
				Expect(service.Save(nil)).To(MatchError("something went wrong"))
			})
		})
	})

	Describe("running a job", func() {
		Context("when the job runs successfully", func() {
			BeforeEach(func() {
				job := jobs.Job{ID: "some-id", Name: "jerb", Command: "doStuff"}
				jobRepo.FindByIdReturns(job, nil)
				runner.RunStub = func(j jobs.Job, oDest io.WriteCloser, status chan<- uint32) error {
					Expect(j).To(Equal(job))
					_, err := oDest.Write([]byte("build output!"))
					Expect(err).NotTo(HaveOccurred())
					Expect(oDest.Close()).To(Succeed())
					status <- 10
					return nil
				}
			})

			It("runs and saves the output of the job", func() {
				r, w := io.Pipe()
				exitCode := make(chan uint32, 1)
				buildRepo.CreateReturns(4, w, exitCode, nil)

				cmdOut := make(chan string)
				go func(c chan<- string) {
					output, err := ioutil.ReadAll(r)
					Expect(err).NotTo(HaveOccurred())
					c <- string(output)
				}(cmdOut)

				buildNumber, err := service.RunJob("some-id")
				Expect(buildNumber).To(Equal(4))
				Expect(err).ToNot(HaveOccurred())
				Expect(runner.RunCallCount()).To(Equal(1))

				Expect(<-cmdOut).To(Equal("build output!"))
				Expect(<-exitCode).To(Equal(uint32(10)))
			})
		})

		Context("when the job cannot be found", func() {
			BeforeEach(func() {
				jobRepo.FindByIdReturns(jobs.Job{}, errors.New("whoops!"))
			})

			It("returns error", func() {
				_, err := service.RunJob("bad-id")
				Expect(err).To(MatchError(ContainSubstring("running job with ID: bad-id")))
			})
		})

		Context("when job cannot be started", func() {
			BeforeEach(func() {
				runner.RunReturns(errors.New("couldn't start job!"))
			})

			It("returns error", func() {
				_, err := service.RunJob("some-id")
				Expect(err).To(MatchError(ContainSubstring("starting job with ID: some-id")))
			})
		})
	})

	Describe("finding a build", func() {
		It("gets the build with complete output from the repository", func() {
			job := jobs.Job{ID: "some-id", Name: "my fancy job"}
			build := jobs.Build{Output: []byte("some output"), ExitStatus: 9, Finished: true}
			buildRepo.FindReturns(build, nil)
			jobRepo.FindByIdReturns(job, nil)

			foundBuild, err := service.FindBuild("some-id", 4)
			Expect(err).NotTo(HaveOccurred())
			Expect(foundBuild).To(Equal(jobs.Build{
				Job:        job,
				Output:     []byte("some output"),
				ExitStatus: 9,
				Finished:   true,
			}))

			Expect(buildRepo.FindCallCount()).To(Equal(1))
			jobId, buildNumber := buildRepo.FindArgsForCall(0)
			Expect(jobId).To(Equal("some-id"))
			Expect(buildNumber).To(Equal(4))
		})

		Context("when finding build fails", func() {
			BeforeEach(func() {
				buildRepo.FindReturns(jobs.Build{}, errors.New("no build here"))
			})

			It("returns the error", func() {
				_, err := service.FindBuild("", 0)
				Expect(err).To(MatchError(ContainSubstring("no build here")))
			})
		})

		Context("when finding build fails", func() {
			BeforeEach(func() {
				jobRepo.FindByIdReturns(jobs.Job{}, errors.New("no build here"))
			})

			It("returns the error", func() {
				_, err := service.FindBuild("", 0)
				Expect(err).To(MatchError(ContainSubstring("no build here")))
			})
		})
	})
})
