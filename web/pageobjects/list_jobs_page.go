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

func (p *ListJobsPage) Visit() *ListJobsPage {
	Expect(p.page.Navigate("http://localhost:3001/jobs")).To(Succeed())
	Eventually(p.page.Find("a#newJob")).Should(BeFound())
	return p
}

func (p *ListJobsPage) GoToCreateNewJob() *NewJobPage {
	Expect(p.page.Find("a#newJob").Click()).To(Succeed())
	Eventually(p.page.Find("form input#name")).Should(BeFound())
	return NewNewJobPage(p.page)
}

func (p *ListJobsPage) GoToBuild(jobName string) *ShowBuildPage {
	Expect(p.page.FindByLink(jobName).Click()).To(Succeed())
	return NewShowBuildPage(p.page)
}
