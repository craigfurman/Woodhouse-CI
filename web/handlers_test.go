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
		handler := web.New(jobService, filepath.Join(cwd, "templates"))
		server = httptest.NewServer(handler)

		page, err = agoutiDriver.NewPage()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(page.Destroy()).To(Succeed())
		server.Close()
	})

	Describe("creating a job", func() {
		It("saves the job", func() {
			By("saving the job using the service", func() {
				jobService.SaveStub = func(job *jobs.Job) error {
					Expect(job.Name).To(Equal("Alice"))
					job.ID = "some-id"
					return nil
				}

				Expect(page.Navigate(fmt.Sprintf("%s/jobs/new", server.URL))).To(Succeed())
				Eventually(page.Find("form input#name")).Should(BeFound())
				Expect(page.Find("form input#name").Fill("Alice")).To(Succeed())
				Expect(page.Find("form button[type=submit]").Click()).To(Succeed())

				Expect(jobService.SaveCallCount()).To(Equal(1))
			})

			By("redirecting to the job output", func() {
				Eventually(page).Should(HaveURL(fmt.Sprintf("%s/jobs/some-id/output", server.URL)))
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
				Eventually(page).Should(HaveURL(fmt.Sprintf("%s/error", server.URL)))
			})
		})
	})

	Describe("job output", func() {
		It("retrieves the job by ID", func() {
			jobService.FindByIdReturns(jobs.Job{Name: "Woodhouse"}, nil)
			Expect(page.Navigate(fmt.Sprintf("%s/jobs/woodhouse-id/output", server.URL))).To(Succeed())
			Eventually(page.Find("#jobTitle")).Should(HaveText("Woodhouse"))

			Expect(jobService.FindByIdCallCount()).To(Equal(1))
			Expect(jobService.FindByIdArgsForCall(0)).To(Equal("woodhouse-id"))
		})

		Context("when retrieving the job fails", func() {
			BeforeEach(func() {
				jobService.FindByIdReturns(jobs.Job{}, errors.New("oops!"))
			})

			It("shows the error page", func() {
				Expect(page.Navigate(fmt.Sprintf("%s/jobs/some-id/output", server.URL))).To(Succeed())
				Eventually(page).Should(HaveURL(fmt.Sprintf("%s/error", server.URL)))
			})
		})
	})
})
