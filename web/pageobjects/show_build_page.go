package pageobjects

import (
	"fmt"

	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

type ShowBuildPage struct {
	page *agouti.Page
}

func NewShowBuildPage(page *agouti.Page) *ShowBuildPage {
	return &ShowBuildPage{page: page}
}

func (p *ShowBuildPage) ScheduleNewBuild() *ShowBuildPage {
	oldUrl, err := p.page.URL()
	Expect(err).NotTo(HaveOccurred())
	Expect(p.page.Find("#startNewBuild").Click()).To(Succeed())
	Eventually(p.page).ShouldNot(HaveURL(oldUrl))
	return p
}

func (p *ShowBuildPage) GoToBuild(buildNumber int) *ShowBuildPage {
	Expect(p.page.FindByLink(fmt.Sprintf("%d", buildNumber)).Click()).To(Succeed())
	return p
}
