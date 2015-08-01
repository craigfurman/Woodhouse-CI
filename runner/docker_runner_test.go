package runner_test

import (
	"errors"
	"path/filepath"

	"github.com/craigfurman/woodhouse-ci/jobs"
	"github.com/craigfurman/woodhouse-ci/runner"
	"github.com/craigfurman/woodhouse-ci/runner/fake_command_runner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DockerRunner", func() {
	var (
		r         *runner.DockerRunner
		cmdRunner *fake_command_runner.FakeCommandRunner
		chunker   runner.ArgChunker

		cmd        = "some command"
		args       = []string{"some", "args have spaces"}
		output     = "some output"
		exitStatus uint32

		rj     jobs.RunningJob
		runErr error
	)

	BeforeEach(func() {
		cmdRunner = new(fake_command_runner.FakeCommandRunner)
		cmdRunner.CombinedOutputReturns([]byte(output), exitStatus, nil)
		chunker = func(argString string) []string {
			Expect(argString).To(Equal(cmd))
			return args
		}
		r = &runner.DockerRunner{
			CommandRunner: cmdRunner,
			ArgChunker:    chunker,
		}
	})

	It("passes the chunked args to the runner", func() {
		Expect(cmdRunner.CombinedOutputCallCount()).To(Equal(1))
		Expect(filepath.Base(cmdRunner.CombinedOutputArgsForCall(0).Path)).To(Equal("docker"))
		Expect(cmdRunner.CombinedOutputArgsForCall(0).Args[1:]).To(ConsistOf("run", "--rm", "busybox", "some", "args have spaces"))
	})

	PContext("when the no arguments can be parsed", func() {
		It("returns error", func() {})
	})

	JustBeforeEach(func() {
		job := jobs.Job{ID: "some-id", Name: "gob", Command: cmd}
		rj, runErr = r.Run(job)
	})

	Context("when the command succeeds", func() {
		BeforeEach(func() {
			exitStatus = 0
		})

		It("does not error", func() {
			Expect(runErr).NotTo(HaveOccurred())
		})

		It("returns combined stdout and stderr and exit status 0", func() {
			Expect(rj).To(Equal(jobs.RunningJob{
				Job:        jobs.Job{ID: "some-id", Name: "gob", Command: cmd},
				Output:     output,
				ExitStatus: 0,
			}))
		})
	})

	Context("when the command returns non-zero exit status", func() {
		BeforeEach(func() {
			exitStatus = 2
		})

		It("does not error", func() {
			Expect(runErr).NotTo(HaveOccurred())
		})

		It("returns combined stdout and stderr and exit status", func() {
			Expect(rj).To(Equal(jobs.RunningJob{
				Job:        jobs.Job{ID: "some-id", Name: "gob", Command: cmd},
				Output:     output,
				ExitStatus: 2,
			}))
		})
	})

	Context("when the command cannot be run", func() {
		BeforeEach(func() {
			cmdRunner.CombinedOutputReturns([]byte{}, 0, errors.New("flan"))
		})

		It("errors", func() {
			Expect(runErr).To(MatchError(ContainSubstring("running command: some command")))
		})
	})
})
