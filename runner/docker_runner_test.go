package runner_test

import (
	"time"

	"github.com/craigfurman/woodhouse-ci/jobs"
	"github.com/craigfurman/woodhouse-ci/runner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("DockerRunner", func() {
	var (
		r *runner.DockerRunner

		cmd    string
		rootFS string

		runErr     error
		output     *gbytes.Buffer
		exitStatus chan uint32
	)

	BeforeEach(func() {
		r = runner.NewDockerRunner()
		output = gbytes.NewBuffer()
		exitStatus = make(chan uint32, 1)
	})

	JustBeforeEach(func() {
		job := jobs.Job{
			ID:          "some-id",
			Name:        "gob",
			Command:     cmd,
			DockerImage: rootFS,
		}
		runErr = r.Run(job, output, exitStatus)
		time.Sleep(time.Second * 2)
	})

	BeforeEach(func() {
		rootFS = "busybox"
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

		It("errors", func() {
			Expect(runErr).To(MatchError(ContainSubstring("running command: somecmd")))
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
