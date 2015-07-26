package runner_test

import (
	"github.com/craigfurman/woodhouse-ci/runner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ArgChunker", func() {
	It("returns empty slice when command is empty string", func() {
		Expect(runner.Chunk("")).To(BeEmpty())
	})

	It("returns slice of single token when command is one token", func() {
		Expect(runner.Chunk("/bin/bash")).To(ConsistOf("/bin/bash"))
	})

	It("returns slice of tokens when command is one token", func() {
		Expect(runner.Chunk("/bin/bash hyphens and-!punctuation")).To(ConsistOf("/bin/bash", "hyphens", "and-!punctuation"))
	})

	It("returns clice of single token when command is single quoted", func() {
		Expect(runner.Chunk(`'one multi-word command'`)).To(ConsistOf("one multi-word command"))
	})

	It("returns clice of single token when command is double quoted", func() {
		Expect(runner.Chunk(`"one multi-word command"`)).To(ConsistOf("one multi-word command"))
	})

	It("returns slice of tokens in complex command", func() {
		Expect(runner.Chunk(`sh -c "echo hi && exit 3"`)).To(ConsistOf("sh", "-c", "echo hi && exit 3"))
	})

	PIt("handles escape characters", func() {})
})
