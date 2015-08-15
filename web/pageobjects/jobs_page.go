package pageobjects

import (
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

type JobsPage struct {
	page *agouti.Page
}

func NewJobsPage(page *agouti.Page) *JobsPage {
	return &JobsPage{page: page}
}

func (p *JobsPage) GoToCreateNewJob() *CreateJobPage {
	Expect(p.page.Find("a#newJob").Click()).To(Succeed())
	Eventually(p.page.Find("form input#name")).Should(BeFound())
	return NewCreateJobPage(p.page)
}
