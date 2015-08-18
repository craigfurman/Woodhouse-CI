package pageobjects

import (
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

type CreateJobPage struct {
	page *agouti.Page
}

func NewCreateJobPage(page *agouti.Page) *CreateJobPage {
	return &CreateJobPage{page: page}
}

func (p *CreateJobPage) CreateJob(name, cmd, dockerImage, gitRepo string) {
	Expect(p.page.Find("form input#name").Fill(name)).To(Succeed())
	Expect(p.page.Find("form input#command").Fill(cmd)).To(Succeed())
	Expect(p.page.Find("form input#dockerImage").Fill(dockerImage)).To(Succeed())
	Expect(p.page.Find("form input#gitRepo").Fill(gitRepo)).To(Succeed())
	Expect(p.page.Find("form button[type=submit]").Click()).To(Succeed())
	Eventually(p.page.Find("#jobTitle")).Should(HaveText(name))
}
