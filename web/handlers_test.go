package web_test

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/craigfurman/woodhouse-ci/jobs"
	"github.com/craigfurman/woodhouse-ci/web"
	"github.com/craigfurman/woodhouse-ci/web/fake_job_service"
	"github.com/craigfurman/woodhouse-ci/web/pageobjects"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

var _ = Describe("Handlers", func() {
	var (
		server *httptest.Server
		page   *agouti.Page

		jobService *fake_job_service.FakeJobService
	)

	BeforeEach(func() {
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		jobService = new(fake_job_service.FakeJobService)
		handler := web.New(jobService, filepath.Join(cwd, "templates"), true)
		server = httptest.NewServer(handler)

		page, err = agoutiDriver.NewPage()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(page.Destroy()).To(Succeed())
		server.Close()
	})

	// TODO is this covered by integration tests?
	PDescribe("listing jobs", func() {})

	Describe("creating a job", func() {
		It("saves the job", func() {
			build := jobs.Build{
				Job:      jobs.Job{Name: "Alice"},
				Output:   []byte("boom!"),
				Finished: true,
			}

			By("saving the job using the service", func() {
				jobService.SaveStub = func(job *jobs.Job) error {
					defer GinkgoRecover()
					Expect(job.Name).To(Equal("Alice"))
					Expect(job.Command).To(Equal("bork bork"))
					Expect(job.DockerImage).To(Equal("user/image:tag"))
					Expect(job.GitRepository).To(Equal("some-repo.git"))
					job.ID = "some-id"
					return nil
				}

				jobService.RunJobReturns(4, nil)
				jobService.FindBuildReturns(build, nil)

				Expect(page.Navigate(fmt.Sprintf("%s/jobs/new", server.URL))).To(Succeed())
				pageobjects.NewNewJobPage(page).CreateJob("Alice", "bork bork", "user/image:tag", "some-repo.git")

				Expect(jobService.SaveCallCount()).To(Equal(1))
			})

			By("redirecting to the build output page", func() {
				Eventually(page).Should(HaveURL(fmt.Sprintf("%s/jobs/some-id/builds/4", server.URL)))
			})

			By("running job", func() {
				Eventually(page.Find("#jobOutput")).Should(HaveText("boom!"))
				Eventually(page.Find("#jobResult")).Should(HaveText("Success"))

				Expect(jobService.RunJobCallCount()).To(Equal(1))
				Expect(jobService.RunJobArgsForCall(0)).To(Equal("some-id"))

				Expect(jobService.FindBuildCallCount()).To(Equal(1))
				jobId, buildNumber := jobService.FindBuildArgsForCall(0)
				Expect(jobId).To(Equal("some-id"))
				Expect(buildNumber).To(Equal(4))
			})
		})

		Context("when saving the job fails", func() {
			BeforeEach(func() {
				jobService.SaveReturns(errors.New("oh dear!"))
			})

			It("shows the error page", func() {
				Expect(page.Navigate(fmt.Sprintf("%s/jobs/new", server.URL))).To(Succeed())
				Eventually(page.Find("form input#name")).Should(BeFound())
				Expect(page.Find("form input#name").Fill("Alice")).To(Succeed())
				Expect(page.Find("form button[type=submit]").Click()).To(Succeed())
				Eventually(page.Find(".errorTrace")).Should(HaveText("oh dear!"))
			})
		})
	})

	Describe("job output", func() {
		Context("when the job is finished", func() {
			It("displays the output", func() {
				By("retrieving the build output", func() {
					jobService.FindBuildReturns(jobs.Build{
						Job:      jobs.Job{Name: "Woodhouse"},
						Output:   []byte("boom!"),
						Finished: true,
					}, nil)
					Expect(page.Navigate(fmt.Sprintf("%s/jobs/woodhouse-id/builds/1", server.URL))).To(Succeed())
					Eventually(page.Find("#jobTitle")).Should(HaveText("Woodhouse"))
					Eventually(page.Find("#jobOutput")).Should(HaveText("boom!"))
					Eventually(page.Find("#jobResult")).Should(HaveText("Success"))

					Expect(jobService.FindBuildCallCount()).To(Equal(1))
					jobId, buildNumber := jobService.FindBuildArgsForCall(0)
					Expect(jobId).To(Equal("woodhouse-id"))
					Expect(buildNumber).To(Equal(1))
				})
			})
		})

		Context("when the job is not finished", func() {
			It("displays the output with pending status", func() {
				By("retrieving the build output", func() {
					jobService.FindBuildReturns(jobs.Build{
						Job:      jobs.Job{Name: "Woodhouse"},
						Output:   []byte("boom!"),
						Finished: false,
					}, nil)
					Expect(page.Navigate(fmt.Sprintf("%s/jobs/woodhouse-id/builds/1", server.URL))).To(Succeed())
					Eventually(page.Find("#jobTitle")).Should(HaveText("Woodhouse"))
					Eventually(page.Find("#jobOutput")).Should(HaveText("boom!"))
					Eventually(page.Find("#jobResult")).Should(HaveText("Running"))

					Expect(jobService.FindBuildCallCount()).To(Equal(1))
					jobId, buildNumber := jobService.FindBuildArgsForCall(0)
					Expect(jobId).To(Equal("woodhouse-id"))
					Expect(buildNumber).To(Equal(1))
				})
			})
		})

		Context("when retrieving the job fails", func() {
			BeforeEach(func() {
				jobService.FindBuildReturns(jobs.Build{}, errors.New("oops!"))
			})

			It("shows the error page", func() {
				Expect(page.Navigate(fmt.Sprintf("%s/jobs/some-id/builds/1", server.URL))).To(Succeed())
				Eventually(page.Find(".errorTrace")).Should(HaveText("oops!"))
			})
		})
	})

	Describe("showing latest build", func() {
		It("redirects to latest build", func() {
			jobService.HighestBuildReturns(42, nil)
			Expect(page.Navigate(fmt.Sprintf("%s/jobs/job-id/builds/latest", server.URL))).To(Succeed())
			Eventually(page).Should(HaveURL(fmt.Sprintf("%s/jobs/job-id/builds/42", server.URL)))

			Expect(jobService.HighestBuildArgsForCall(0)).To(Equal("job-id"))
		})
	})
})
