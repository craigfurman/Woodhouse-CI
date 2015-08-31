package integration_test

import (
	"strings"
	"time"

	"github.com/craigfurman/woodhouse-ci/web/pageobjects"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

var _ = Describe("Creating a job", func() {
	var (
		page *agouti.Page

		repo = "https://github.com/craigfurman/Woodhouse-CI.git"
	)

	BeforeEach(func() {
		var err error
		page, err = agoutiDriver.NewPage()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(page.Destroy()).To(Succeed())
	})

	PContext("when there is no docker image specified", func() {
		It("runs the job on the host OS", func() {})
	})

	It("creates and runs the new job", func() {
		By("creating the new job using the specified docker image", func() {
			pageobjects.NewListJobsPage(page).Visit().
				GoToCreateNewJob().
				CreateJob("Bob", "cat /etc/lsb-release", "ubuntu:14.04.3", "")
		})

		By("streaming the output from the job", func() {
			expected := `.*DISTRIB_ID=Ubuntu
DISTRIB_RELEASE=14.04
DISTRIB_CODENAME=trusty
DISTRIB_DESCRIPTION="Ubuntu 14.04.3 LTS".*`
			Eventually(page.Find("#jobOutput")).Should(MatchText(expected))
		})

		By("indicating that the job ran successfully", func() {
			Eventually(page.Find("#jobResult")).Should(HaveText("Success"))
		})
	})

	Context("when 2 builds are scheduled for a job", func() {
		It("preserves build history", func() {
			var showBuildPage *pageobjects.ShowBuildPage
			By("creating the new job", func() {
				showBuildPage = pageobjects.NewListJobsPage(page).Visit().
					GoToCreateNewJob().
					CreateJob("busyJob", "echo hello", "busybox", "")
			})

			var latestBuildUrl string
			By("scheduling another build", func() {
				oldUrl, err := page.URL()
				Expect(err).NotTo(HaveOccurred())

				showBuildPage.ScheduleNewBuild()

				parts := strings.Split(oldUrl, "/")
				parts = parts[0 : len(parts)-1]
				parts = append(parts, "2")
				latestBuildUrl = strings.Join(parts, "/")
				Eventually(page).Should(HaveURL(latestBuildUrl))
			})

			By("linking to the latest build on the list jobs page", func() {
				pageobjects.NewListJobsPage(page).Visit().GoToBuild("busyJob")
				Eventually(page).Should(HaveURL(latestBuildUrl))
			})
		})
	})

	Describe("specifying a git repository for the job", func() {
		It("makes the repository available at the specified relative path", func() {
			By("mounting the repository inside the container", func() {
				pageobjects.NewListJobsPage(page).Visit().
					GoToCreateNewJob().
					CreateJob("Bob", "cat LICENSE", "busybox", repo)
			})

			By("streaming the output from the job", func() {
				Eventually(page.Find("#jobOutput")).Should(MatchText(".*Craig Furman.*"))
			})
		})
	})

	Context("when the job takes a long time to complete", func() {
		It("starts the job and streams the output", func() {
			By("creating the new job", func() {
				jobSubmitted := make(chan bool)
				go func(c chan<- bool) {
					defer GinkgoRecover()
					pageobjects.NewListJobsPage(page).Visit().
						GoToCreateNewJob().
						CreateJob("Bob", `sh -c "echo start && sleep 2 && echo finish"`, "busybox", "")
					c <- true
				}(jobSubmitted)
				select {
				case <-jobSubmitted:
				case <-time.After(time.Second * 2):
					Fail("timed out waiting for job to be submitted")
				}
			})

			By("streaming the output from the job", func() {
				Eventually(page.Find("#jobOutput")).Should(HaveText("start"))
				Eventually(page.Find("#jobOutput"), "5s").Should(HaveText("start\nfinish"))
			})

			By("indicating that the job ran successfully", func() {
				Eventually(page.Find("#jobResult")).Should(HaveText("Success"))
			})
		})
	})

	Context("when the job fails", func() {
		It("creates and runs the new job", func() {
			By("creating the new job", func() {
				pageobjects.NewListJobsPage(page).Visit().
					GoToCreateNewJob().
					CreateJob("FailedJob", `sh -c "echo hi >&2 && exit 42"`, "busybox", "")
			})

			By("streaming the output from the job", func() {
				Eventually(page.Find("#jobOutput")).Should(HaveText("hi"))
			})

			By("indicating that the job ran successfully", func() {
				Eventually(page.Find("#jobResult")).Should(HaveText("Failure: exit status 42"))
			})
		})
	})
})
