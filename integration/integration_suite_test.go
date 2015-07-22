package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/agouti"

	"testing"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var (
	executablePath    string
	runningExecutable *gexec.Session

	agoutiDriver *agouti.WebDriver
)

var _ = BeforeSuite(func() {
	var err error
	executablePath, err = gexec.Build("github.com/craigfurman/woodhouse-ci")
	Expect(err).NotTo(HaveOccurred())

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

var _ = BeforeEach(func() {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	templateDir := filepath.Join(cwd, "..", "web", "templates")
	runningExecutable, err = gexec.Start(exec.Command(executablePath, "-port=3000", "-templateDir", templateDir), GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterEach(func() {
	Eventually(runningExecutable.Kill()).Should(gexec.Exit())
})
