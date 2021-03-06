package integration_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/agouti"

	"testing"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(time.Second * 5)
	RunSpecs(t, "Integration Suite")
}

var (
	executablePath    string
	runningExecutable *gexec.Session
	buildsDir         string
	agoutiDriver      *agouti.WebDriver
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
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	storeDir := filepath.Join(cwd, "..", "db")
	os.Remove(filepath.Join(storeDir, "sqlite", "store.db"))

	buildsDir, err = ioutil.TempDir("", "integration-builds")
	Expect(err).NotTo(HaveOccurred())

	runningExecutable, err = gexec.Start(exec.Command(
		executablePath, "-port=3001",
		"-templateDir", filepath.Join(cwd, "..", "web", "templates"),
		"-storeDir", storeDir,
		"-gooseCmd=goose",
		"-buildsDir", buildsDir,
		"-assetsDir", filepath.Join(cwd, "..", "web", "assets"),
	), GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterEach(func() {
	Eventually(runningExecutable.Kill()).Should(gexec.Exit())
	Expect(os.RemoveAll(buildsDir)).To(Succeed())
})
