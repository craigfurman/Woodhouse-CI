package integration_test

import (
	"time"

	"github.com/craigfurman/woodhouse-ci/web/pageobjects"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

var _ = Describe("Creating a job", func() {
	var page *agouti.Page

	BeforeEach(func() {
		var err error
		page, err = agoutiDriver.NewPage()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(page.Destroy()).To(Succeed())
	})

	It("creates and runs the new job", func() {
		By("navigating to the jobs page", func() {
			Expect(page.Navigate("http://localhost:3001/jobs")).To(Succeed())
			Eventually(page.Find("a#newJob")).Should(BeFound())
		})

		By("creating the new job", func() {
			pageobjects.NewJobsPage(page).GoToCreateNewJob().CreateJob("Bob", "echo good morning")
		})

		By("streaming the output from the job", func() {
			Eventually(page.Find("#jobOutput")).Should(HaveText("good morning"))
		})

		By("indicating that the job ran successfully", func() {
			Eventually(page.Find("#jobResult")).Should(HaveText("Success"))
		})
	})

	Context("when the job takes a long time to complete", func() {
		It("starts the job and streams the output", func() {
			By("navigating to the jobs page", func() {
				Expect(page.Navigate("http://localhost:3001/jobs")).To(Succeed())
				Eventually(page.Find("a#newJob")).Should(BeFound())
			})

			By("creating the new job", func() {
				jobSubmitted := make(chan bool)
				go func(c chan<- bool) {
					defer GinkgoRecover()
					pageobjects.NewJobsPage(page).GoToCreateNewJob().CreateJob("Bob", `sh -c "echo start && sleep 4 && echo finish"`)
					c <- true
				}(jobSubmitted)
				select {
				case <-jobSubmitted:
				case <-time.After(time.Second):
					Fail("timed out waiting for job to be submitted")
				}
			})

			By("streaming the output from the job", func() {
				Eventually(page.Find("#jobOutput")).Should(HaveText("start"))
				Eventually(page.Find("#jobOutput"), "5s").Should(HaveText("start\nfinish"))
			})

			By("indicating that the job ran successfully", func() {
				Eventually(page.Find("#jobResult")).Should(HaveText("Success"))
			})
		})
	})

	Context("when the job fails", func() {
		It("creates and runs the new job", func() {
			By("navigating to the jobs page", func() {
				Expect(page.Navigate("http://localhost:3001/jobs")).To(Succeed())
				Eventually(page.Find("a#newJob")).Should(BeFound())
			})

			By("creating the new job", func() {
				pageobjects.NewJobsPage(page).GoToCreateNewJob().CreateJob("FailedJob", `sh -c "echo hi >&2 && exit 42"`)
			})

			By("streaming the output from the job", func() {
				Eventually(page.Find("#jobOutput")).Should(HaveText("hi"))
			})

			By("indicating that the job ran successfully", func() {
				Eventually(page.Find("#jobResult")).Should(HaveText("Failure: exit status 42"))
			})
		})
	})
})
