package builds_test

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/craigfurman/woodhouse-ci/builds"
	"github.com/craigfurman/woodhouse-ci/jobs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildRepository", func() {
	var (
		repo      *builds.Repository
		buildsDir string
	)

	Describe("Creating a new build", func() {
		var (
			buildNumber    int
			outputDest     io.WriteCloser
			exitStatusChan chan uint32
			err            error
		)

		JustBeforeEach(func() {
			buildNumber, outputDest, exitStatusChan, err = repo.Create("some-id")
		})

		Context("when the builds directory already exists", func() {
			BeforeEach(func() {
				var err error
				buildsDir, err = ioutil.TempDir("", "builds")
				Expect(err).NotTo(HaveOccurred())
				repo = &builds.Repository{BuildsDir: buildsDir}
			})

			AfterEach(func() {
				Expect(os.RemoveAll(buildsDir)).To(Succeed())
			})

			It("does not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when output and exitStatus are written", func() {
				JustBeforeEach(func() {
					_, err := outputDest.Write([]byte("output from build"))
					Expect(err).NotTo(HaveOccurred())
					exitStatusChan <- 42
					Eventually(func() error {
						_, err := os.Stat(filepath.Join(buildsDir, "some-id", "1-status.txt"))
						return err
					}).ShouldNot(HaveOccurred())
				})

				Describe("retrieving the build info", func() {
					var (
						b       jobs.Build
						findErr error
					)

					JustBeforeEach(func() {
						b, findErr = repo.Find("some-id", buildNumber)
					})

					Context("when the build is finished", func() {
						It("does not error", func() {
							Expect(findErr).NotTo(HaveOccurred())
						})

						It("returns the build info", func() {
							Expect(b.Output).To(Equal("output from build"))
							Expect(b.ExitStatus).To(Equal(uint32(42)))
						})
					})

					Context("when no builds exist for the given Job", func() {
						It("returns error", func() {
							_, err := repo.Find("idontexist", 1)
							Expect(err).To(MatchError(ContainSubstring("no builds found for job idontexist")))
						})
					})
				})

				Context("when another build is created", func() {
					JustBeforeEach(func() {
						_, _, _, err := repo.Create("some-other-id")
						Expect(err).NotTo(HaveOccurred())
					})

					It("can still find the previous build data", func() {
						b, err := repo.Find("some-id", buildNumber)
						Expect(err).NotTo(HaveOccurred())
						Expect(b.Output).To(Equal("output from build"))
						Expect(b.ExitStatus).To(Equal(uint32(42)))
					})
				})
			})
		})

		Context("when the builds directory does not already exist", func() {
			It("does not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
