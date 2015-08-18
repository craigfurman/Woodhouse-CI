package runner_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/craigfurman/woodhouse-ci/jobs"
	"github.com/craigfurman/woodhouse-ci/runner"
	"github.com/craigfurman/woodhouse-ci/runner/fake_vcs_fetcher"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("DockerRunner", func() {
	var (
		r          *runner.DockerRunner
		vcsFetcher *fake_vcs_fetcher.FakeVcsFetcher

		cmd           string
		rootFS        string
		gitRepository string

		runErr     error
		output     *gbytes.Buffer
		exitStatus chan uint32
	)

	BeforeEach(func() {
		vcsFetcher = new(fake_vcs_fetcher.FakeVcsFetcher)
		r = runner.NewDockerRunner(vcsFetcher)
		output = gbytes.NewBuffer()
		exitStatus = make(chan uint32, 1)
	})

	JustBeforeEach(func() {
		job := jobs.Job{
			ID:            "some-id",
			Name:          "gob",
			Command:       cmd,
			DockerImage:   rootFS,
			GitRepository: gitRepository,
		}
		runErr = r.Run(job, output, exitStatus)
		time.Sleep(time.Second * 2)
	})

	BeforeEach(func() {
		rootFS = "busybox"
		gitRepository = ""
	})

	Context("when the command succeeds", func() {
		BeforeEach(func() {
			cmd = "echo hello"
		})

		It("does not error", func() {
			Expect(runErr).NotTo(HaveOccurred())
		})

		It("asynchronously writes combined stdout and stderr", func() {
			Eventually(output).Should(gbytes.Say("hello"))
		})

		It("sends the status code", func() {
			Expect(<-exitStatus).To(Equal(uint32(0)))
		})

		It("closes the output writer", func() {
			Eventually(output.Closed()).Should(BeTrue())
		})

		Context("and the job has no git repository", func() {
			It("does not fetch from any repository", func() {
				Expect(vcsFetcher.FetchCallCount()).To(Equal(0))
			})
		})

		Context("and the job has a git repository", func() {
			var repoDir string

			BeforeEach(func() {
				var err error
				repoDir, err = ioutil.TempDir("", "docker-runner-unit-tests")
				Expect(err).NotTo(HaveOccurred())
				f, err := os.Create(filepath.Join(repoDir, "test.txt"))
				Expect(err).NotTo(HaveOccurred())
				_, err = f.Write([]byte("hello from tests!"))
				Expect(err).NotTo(HaveOccurred())
				Expect(f.Close()).To(Succeed())

				vcsFetcher.FetchReturns(repoDir, nil)
				gitRepository = "some-repo"
				cmd = "cat test.txt"
			})

			itRemovesTheRepo := func() {
				It("removes the repo", func() {
					Eventually(func() bool {
						_, err := os.Stat(repoDir)
						return os.IsNotExist(err)
					}).Should(BeTrue())
				})
			}

			itRemovesTheRepo()

			It("runs the job with the repo mounted in the container as cwd", func() {
				Eventually(output).Should(gbytes.Say("hello from tests!"))
				repo, _ := vcsFetcher.FetchArgsForCall(0)
				Expect(repo).To(Equal("some-repo"))
			})

			Context("when fetching fails", func() {
				BeforeEach(func() {
					vcsFetcher.FetchReturns(repoDir, errors.New("oops"))
				})

				It("does not error", func() {
					Expect(runErr).NotTo(HaveOccurred())
				})

				It("sends exit status 1", func() {
					Expect(<-exitStatus).To(Equal(uint32(1)))
				})

				itRemovesTheRepo()
			})
		})

		Describe("the docker image for the job", func() {
			BeforeEach(func() {
				rootFS = "debian:jessie"
				cmd = "cat /etc/os-release"
			})

			It("runs the job using the specified docker image", func() {
				<-exitStatus
				Expect(string(output.Contents())).To(ContainSubstring("Debian GNU/Linux 8 (jessie)"))
			})
		})
	})

	Context("when the command returns non-zero exit status", func() {
		BeforeEach(func() {
			cmd = `sh -c "echo hello && exit 2"`
		})

		It("asynchronously writes combined stdout and stderr", func() {
			Eventually(output).Should(gbytes.Say("hello"))
		})

		It("sends the status code", func() {
			Expect(<-exitStatus).To(Equal(uint32(2)))
		})
	})

	Context("when the docker image is not specified", func() {
		BeforeEach(func() {
			cmd = "echo hello"
			rootFS = ""
		})

		It("errors", func() {
			Expect(runErr).To(MatchError("you need to specify a docker image when using DockerRunner"))
		})
	})

	Context("when the docker image does not exist", func() {
		BeforeEach(func() {
			cmd = "echo hello"
			rootFS = "WoodhouseOS:#notarealthing"
		})

		It("returns non-zero exit status", func() {
			Expect(<-exitStatus).ToNot(Equal(uint32(0)))
		})
	})

	Context("when the command cannot be run", func() {
		BeforeEach(func() {
			r.DockerCmd = "ihopethisdoesntexistonpath"
			cmd = "somecmd"
		})

		It("sends exit status 1 to represent failure to fork", func() {
			Expect(<-exitStatus).To(Equal(uint32(1)))
		})
	})

	Context("when the no arguments can be parsed", func() {
		BeforeEach(func() {
			cmd = ""
		})

		It("returns error", func() {
			Expect(runErr).To(MatchError("No arguments could be parsed from command: "))
		})
	})
})
