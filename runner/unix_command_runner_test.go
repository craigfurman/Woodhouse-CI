package runner_test

import (
	"os/exec"

	"github.com/craigfurman/woodhouse-ci/runner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UnixCommandRunner", func() {
	var (
		r = runner.UnixCommandRunner{}

		cmd *exec.Cmd

		output     []byte
		exitStatus uint32
		runErr     error
	)

	JustBeforeEach(func() {
		output, exitStatus, runErr = r.CombinedOutput(cmd)
	})

	Context("when the command exits with 0", func() {
		BeforeEach(func() {
			cmd = exec.Command("sh", "-c", `echo "hello from stdout" && echo "hello from stderr" >&2`)
		})

		It("returns combined stdout and stderr", func() {
			Expect(string(output)).To(Equal("hello from stdout\nhello from stderr\n"))
		})

		It("returns exit status 0", func() {
			Expect(exitStatus).To(Equal(uint32(0)))
		})

		It("does not error", func() {
			Expect(runErr).NotTo(HaveOccurred())
		})
	})

	Context("when the command does not exit with 0", func() {
		BeforeEach(func() {
			cmd = exec.Command("sh", "-c", "echo hi && exit 3")
		})

		It("returns combined stdout and stderr", func() {
			Expect(string(output)).To(Equal("hi\n"))
		})

		It("returns exit status", func() {
			Expect(exitStatus).To(Equal(uint32(3)))
		})

		It("does not error", func() {
			Expect(runErr).NotTo(HaveOccurred())
		})
	})

	Context("when the command cannot be executed", func() {
		BeforeEach(func() {
			cmd = exec.Command("ireallydoubtthisisinpath")
		})

		It("returns the error", func() {
			Expect(runErr).To(HaveOccurred())
		})
	})
})
