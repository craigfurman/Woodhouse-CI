package integration_test

import (
	"github.com/craigfurman/woodhouse-ci/web/pageobjects"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

var _ = Describe("ListJobs", func() {
	var page *agouti.Page

	BeforeEach(func() {
		var err error
		page, err = agoutiDriver.NewPage()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(page.Destroy()).To(Succeed())
	})

	It("shows a list of jobs", func() {
		By("root url redirects to jobs page", func() {
			Expect(page.Navigate("http://localhost:3001")).To(Succeed())
			Eventually(page.Find("a#newJob")).Should(BeFound())
		})

		By("creating a new job", func() {
			pageobjects.NewListJobsPage(page).GoToCreateNewJob().CreateJob("Jerb", "echo hello", "busybox", "")
		})

		By("list includes the job on jobs page", func() {
			Expect(page.Navigate("http://localhost:3001/jobs")).To(Succeed())
			Eventually(page.Find(".job:first-of-type")).Should(MatchText(".*Jerb.*"))
		})
	})
})
