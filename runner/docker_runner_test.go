package runner_test

import (
	"io"
	"io/ioutil"

	"github.com/craigfurman/woodhouse-ci/jobs"
	"github.com/craigfurman/woodhouse-ci/runner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DockerRunner", func() {
	var (
		r *runner.DockerRunner

		cmd string

		runErr     error
		output     chan string
		exitStatus chan uint32
	)

	BeforeEach(func() {
		r = runner.NewDockerRunner()
		output = make(chan string, 1)
		exitStatus = make(chan uint32, 1)
	})

	PContext("when the no arguments can be parsed", func() {
		It("returns error", func() {})
	})

	JustBeforeEach(func() {
		job := jobs.Job{ID: "some-id", Name: "gob", Command: cmd}
		reader, w := io.Pipe()
		go func(c chan<- string) {
			b, err := ioutil.ReadAll(reader)
			Expect(err).NotTo(HaveOccurred())
			c <- string(b)
		}(output)

		runErr = r.Run(job, w, exitStatus)
	})

	Context("when the command succeeds", func() {
		BeforeEach(func() {
			cmd = "echo hello"
		})

		It("does not error", func() {
			Expect(runErr).NotTo(HaveOccurred())
		})

		It("synchronously writes combined stdout and stderr", func() {
			Expect(<-output).To(Equal("hello\n"))
		})

		It("sends the status code", func() {
			Expect(<-exitStatus).To(Equal(uint32(0)))
		})
	})

	Context("when the command returns non-zero exit status", func() {
		BeforeEach(func() {
			cmd = `sh -c "echo hello && exit 2"`
		})

		It("synchronously writes combined stdout and stderr", func() {
			Expect(<-output).To(Equal("hello\n"))
		})

		It("sends the status code", func() {
			Expect(<-exitStatus).To(Equal(uint32(2)))
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
})
