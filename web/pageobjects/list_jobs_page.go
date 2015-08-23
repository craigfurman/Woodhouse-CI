package pageobjects

import (
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

type ListJobsPage struct {
	page *agouti.Page
}

func NewListJobsPage(page *agouti.Page) *ListJobsPage {
	return &ListJobsPage{page: page}
}

func (p *ListJobsPage) GoToCreateNewJob() *NewJobPage {
	Expect(p.page.Find("a#newJob").Click()).To(Succeed())
	Eventually(p.page.Find("form input#name")).Should(BeFound())
	return NewNewJobPage(p.page)
}
