package web_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"

	"testing"
)

func TestWeb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Web Suite")
}

var agoutiDriver *agouti.WebDriver

var _ = BeforeSuite(func() {
	if os.Getenv("HEADLESS") == "true" {
		agoutiDriver = agouti.PhantomJS()
	} else {
		agoutiDriver = agouti.ChromeDriver()
	}

	Expect(agoutiDriver.Start()).To(Succeed())
})

var _ = AfterSuite(func() {
	Expect(agoutiDriver.Stop()).To(Succeed())
})
