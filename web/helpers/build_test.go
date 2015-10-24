package helpers_test

import (
	"html/template"

	"github.com/craigfurman/woodhouse-ci/jobs"
	"github.com/craigfurman/woodhouse-ci/web/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Build view helpers", func() {
	Describe("sanitized output", func() {
		It("replaces newlines with html <br>", func() {
			Expect(helpers.SanitisedHTML([]byte("some\nlines"))).To(Equal(template.HTML("some<br>lines")))
		})

		FIt("replaces ANSI colour escape codes with styles", func() {
			Expect(helpers.SanitisedHTML([]byte(`\e[95msome colour text \e[94ma different colour`))).To(Equal(template.HTML(`<span class="ansi-95">some colour text </span><span class="ansi-94">a different colour</span>`)))
		})

		It("still escapes all HTML in output", func() {
			Expect(helpers.SanitisedHTML([]byte("<script>dangerous</script>"))).To(Equal(template.HTML("&lt;script&gt;dangerous&lt;/script&gt;")))
		})
	})

	Describe("exit message", func() {
		It("returns success when exit status is 0", func() {
			Expect(helpers.Message(jobs.Build{
				Finished:   true,
				ExitStatus: 0,
			})).To(Equal("Success"))
		})

		It("returns failure when exit status is non-zero", func() {
			Expect(helpers.Message(jobs.Build{
				Finished:   true,
				ExitStatus: 42,
			})).To(Equal("Failure: exit status 42"))
		})

		It("returns running when the build is not finished", func() {
			Expect(helpers.Message(jobs.Build{
				Finished: false,
			})).To(Equal("Running"))
		})
	})

	Describe("build classes", func() {
		var (
			b       jobs.Build
			classes string
		)

		JustBeforeEach(func() {
			classes = helpers.Classes(b)
		})

		Context("when the build is running", func() {
			BeforeEach(func() {
				b = jobs.Build{Finished: false}
			})

			It("returns empty string", func() {
				Expect(classes).To(BeEmpty())
			})
		})

		Context("when the build was successful", func() {
			BeforeEach(func() {
				b = jobs.Build{Finished: true, ExitStatus: 0}
			})

			It("returns empty string", func() {
				Expect(classes).To(Equal("passing"))
			})
		})

		Context("when the build has failed", func() {
			BeforeEach(func() {
				b = jobs.Build{Finished: true, ExitStatus: 1}
			})

			It("returns empty string", func() {
				Expect(classes).To(Equal("failing"))
			})
		})
	})
})
