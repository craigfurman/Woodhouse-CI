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

		By("indicating that the job is passing on job list page", func() {
			pageobjects.NewListJobsPage(page).Visit()
			Eventually(page.FindByLink("Bob")).Should(HaveAttribute("class", "passing"))
		})
	})

	FIt("streams output in colour", func() {
		pageobjects.NewListJobsPage(page).Visit().GoToCreateNewJob().
			CreateJob("colour output!", `echo -e "\e[96mSuch colour!"`, "busybox", "")

		Eventually(page.Find("#jobOutput")).Should(HaveText("Such colour!"))
		Eventually(page.Find("#jobOutput")).ShouldNot(MatchText(`.*\[.*`))
	})

	Context("when 2 builds are scheduled for a job", func() {
		It("preserves build history", func() {
			var showBuildPage *pageobjects.ShowBuildPage
			By("creating the new job", func() {
				showBuildPage = pageobjects.NewListJobsPage(page).Visit().
					GoToCreateNewJob().
					CreateJob("busyJob", "echo hello", "busybox", "")
			})

			var firstBuildURL, latestBuildURL string
			By("scheduling another build", func() {
				var err error
				firstBuildURL, err = page.URL()
				Expect(err).NotTo(HaveOccurred())

				showBuildPage.ScheduleNewBuild()

				parts := strings.Split(firstBuildURL, "/")
				parts = parts[0 : len(parts)-1]
				parts = append(parts, "2")
				latestBuildURL = strings.Join(parts, "/")
				Eventually(page).Should(HaveURL(latestBuildURL))
			})

			By("linking to the latest build on the list jobs page", func() {
				showBuildPage = pageobjects.NewListJobsPage(page).Visit().GoToBuild("busyJob")
				Eventually(page).Should(HaveURL(latestBuildURL))
			})

			By("showing the build history", func() {
				showBuildPage.GoToBuild(1)
				Eventually(page).Should(HaveURL(firstBuildURL))
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

			By("updating the jobs status on the list jobs page", func() {
				pageobjects.NewShowBuildPage(page).ScheduleNewBuild()
				pageobjects.NewListJobsPage(page).Visit()
				Eventually(page.FindByLink("Bob")).Should(HaveAttribute("class", "passing"))
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

			By("indicating that the job failed on the job output page", func() {
				Eventually(page.Find("#jobResult")).Should(HaveText("Failure: exit status 42"))
			})

			By("indicating that the job is failing on job list page", func() {
				pageobjects.NewListJobsPage(page).Visit()
				Eventually(page.FindByLink("FailedJob")).Should(HaveAttribute("class", "failing"))
			})
		})
	})
})
