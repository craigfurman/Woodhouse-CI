package builds_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/craigfurman/woodhouse-ci/blockingio"
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
			jobId = "some-id"

			buildNumber    int
			outputDest     io.WriteCloser
			exitStatusChan chan uint32
			createErr      error
		)

		JustBeforeEach(func() {
			buildNumber, outputDest, exitStatusChan, createErr = repo.Create(jobId)
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
				Expect(createErr).NotTo(HaveOccurred())
			})

			Context("when output and exitStatus are written", func() {
				JustBeforeEach(func() {
					_, err := outputDest.Write([]byte("output from build"))
					Expect(err).NotTo(HaveOccurred())
					Expect(outputDest.Close()).To(Succeed())
					exitStatusChan <- 42
					Eventually(func() error {
						_, err := os.Stat(filepath.Join(buildsDir, jobId, "1-status.txt"))
						return err
					}).ShouldNot(HaveOccurred())
				})

				Describe("retrieving the build info", func() {
					var (
						b       jobs.Build
						findErr error
					)

					JustBeforeEach(func() {
						b, findErr = repo.Find(jobId, buildNumber)
					})

					It("does not error", func() {
						Expect(findErr).NotTo(HaveOccurred())
					})

					It("returns the build info", func() {
						Expect(b.Output).To(Equal([]byte("output from build")))
						Expect(b.ExitStatus).To(Equal(uint32(42)))
						Expect(b.Finished).To(BeTrue())
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
						b, err := repo.Find(jobId, buildNumber)
						Expect(err).NotTo(HaveOccurred())
						Expect(b.Output).To(Equal([]byte("output from build")))
						Expect(b.ExitStatus).To(Equal(uint32(42)))
						Expect(b.Finished).To(BeTrue())
					})
				})
			})

			Context("when output is being written but the job is not finished", func() {
				JustBeforeEach(func() {
					_, err := outputDest.Write([]byte("output from build"))
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					outputDest.Close()
				})

				Describe("retrieving the build info", func() {
					var (
						b       jobs.Build
						findErr error
					)

					JustBeforeEach(func() {
						b, findErr = repo.Find(jobId, buildNumber)
					})

					It("does not error", func() {
						Expect(findErr).NotTo(HaveOccurred())
					})

					It("returns the build info", func() {
						Expect(b.Output).To(Equal([]byte("output from build")))
						Expect(b.Finished).To(BeFalse())
					})
				})

				Describe("streaming output from the build", func() {
					var (
						streamer *blockingio.BlockingReader
						sErr     error

						jobIdToStream       string
						buildNumberToStream int
						startAtByte         int64
					)

					BeforeEach(func() {
						jobIdToStream = jobId
						buildNumberToStream = 1
						startAtByte = 0
					})

					JustBeforeEach(func() {
						streamer, sErr = repo.Stream(jobIdToStream, buildNumberToStream, startAtByte)
					})

					Context("when the jobId and build number are valid", func() {
						AfterEach(func() {
							if streamer != nil {
								Expect(streamer.Close()).To(Succeed())
							}
						})

						It("does not error", func() {
							Expect(buildNumber).To(Equal(1))
							Expect(buildNumberToStream).To(Equal(1))
							Expect(sErr).NotTo(HaveOccurred())
						})

						It("blocks on read operations", func() {
							done := make(chan bool)
							go func(c chan<- bool) {
								defer GinkgoRecover()
								content := []byte{}
								for {
									nextBytes, done := streamer.Next()
									content = append(content, nextBytes...)
									if done {
										break
									}
								}
								Expect(content).To(Equal([]byte("output from builda linemore\nlines")))
								c <- true
							}(done)

							_, err := outputDest.Write([]byte("a line"))
							Expect(err).NotTo(HaveOccurred())
							_, err = outputDest.Write([]byte("more\nlines"))
							Expect(err).NotTo(HaveOccurred())
							Expect(outputDest.Close()).To(Succeed())
							exitStatusChan <- 0

							select {
							case <-done:
							case <-time.After(time.Second * 5):
								Fail("timed out")
							}
						})

						Context("when the offset to start at is non-zero", func() {
							BeforeEach(func() {
								startAtByte = int64(len([]byte("output from build")))
							})

							It("blocks on read operations when streaming the output", func() {
								_, err := outputDest.Write([]byte("further output"))
								Expect(err).NotTo(HaveOccurred())
								b, _ := streamer.Next()
								// Expect(done).To(BeTrue())
								Expect(b).To(Equal([]byte("further output")))
							})

							Context("when the offset is negative", func() {
								BeforeEach(func() {
									startAtByte = -1
								})

								It("errors", func() {
									Expect(sErr).To(MatchError(ContainSubstring("seeking")))
								})
							})
						})

						PContext("when the output cannot be read", func() {
							It("errors", func() {})
						})
					})

					Context("when thejob does not exist", func() {
						BeforeEach(func() {
							jobIdToStream = "idontexist"
						})

						It("errors", func() {
							Expect(sErr).To(MatchError(ContainSubstring(fmt.Sprintf("streaming output from job: idontexist, build: %d", buildNumberToStream))))
						})
					})

					Context("when the build number does not exist", func() {
						BeforeEach(func() {
							buildNumberToStream = -1
						})

						It("errors", func() {
							Expect(sErr).To(MatchError(ContainSubstring("streaming output from job: some-id, build: -1")))
						})
					})
				})
			})
		})

		Context("when the builds directory does not already exist", func() {
			BeforeEach(func() {
				tmpDir, err := ioutil.TempDir("", "build-repo-unit-tests")
				Expect(err).NotTo(HaveOccurred())
				buildsDir = filepath.Join(tmpDir, "i-dont-exist-yet")
				repo = &builds.Repository{BuildsDir: buildsDir}
			})

			AfterEach(func() {
				Expect(os.RemoveAll(buildsDir)).To(Succeed())
			})

			It("does not error", func() {
				Expect(createErr).NotTo(HaveOccurred())
			})

			It("creates it", func() {
				_, err := os.Stat(buildsDir)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
