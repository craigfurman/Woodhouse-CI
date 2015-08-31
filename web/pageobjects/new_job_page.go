package pageobjects

import (
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

type NewJobPage struct {
	page *agouti.Page
}

func NewNewJobPage(page *agouti.Page) *NewJobPage {
	return &NewJobPage{page: page}
}

func (p *NewJobPage) CreateJob(name, cmd, dockerImage, gitRepo string) *ShowBuildPage {
	Expect(p.page.Find("form input#name").Fill(name)).To(Succeed())
	Expect(p.page.Find("form input#command").Fill(cmd)).To(Succeed())
	Expect(p.page.Find("form input#dockerImage").Fill(dockerImage)).To(Succeed())
	Expect(p.page.Find("form input#gitRepo").Fill(gitRepo)).To(Succeed())
	Expect(p.page.Find("form button[type=submit]").Click()).To(Succeed())
	Eventually(p.page.Find("#jobTitle")).Should(HaveText(name))
	return NewShowBuildPage(p.page)
}
