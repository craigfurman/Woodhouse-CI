package integration_test

import (
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
			Expect(page.Navigate("http://localhost:3000/jobs")).To(Succeed())
			Eventually(page.Find("a#newJob")).Should(BeFound())
		})

		By("creating the new job", func() {
			Expect(page.Find("a#newJob").Click()).To(Succeed())
			Eventually(page.Find("form input#name")).Should(BeFound())
			Expect(page.Find("form input#name").Fill("Bob")).To(Succeed())
			Expect(page.Find("form button[type=submit]").Click()).To(Succeed())
			Eventually(page.Find("h1#jobTitle")).Should(HaveText("Bob"))
		})

		By("streaming the output from the job", func() {
			Expect(page.Find("#jobOutput")).To(HaveText("Hello world!"))
		})
	})
})
