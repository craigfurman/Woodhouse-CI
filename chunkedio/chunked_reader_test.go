package chunkedio_test

import (
	"bytes"
	"io"
	"time"

	"github.com/craigfurman/woodhouse-ci/chunkedio"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ChunkedReader", func() {
	var (
		br     *chunkedio.ChunkedReader
		writer *io.PipeWriter

		done bool
	)

	BeforeEach(func() {
		var reader *io.PipeReader
		reader, writer = io.Pipe()
		br = &chunkedio.ChunkedReader{
			Output:      reader,
			DoneWriting: func() bool { return done },
			Buffer:      make([]byte, 1024),
		}
	})

	Context("when the writer is already done and no output has been written", func() {
		BeforeEach(func() {
			Expect(writer.Close()).To(Succeed())
			done = true
		})

		It("returns empty slice and true", func() {
			out, finished := br.Next()
			Expect(out).To(BeEmpty())
			Expect(finished).To(BeTrue())
		})
	})

	Context("when there is a small amount of output and the writer is done", func() {
		var testFinished chan bool

		BeforeEach(func() {
			testFinished = make(chan bool)

			go func() {
				defer GinkgoRecover()
				_, err := writer.Write([]byte("Hello world!"))
				Expect(err).NotTo(HaveOccurred())
				Expect(writer.Close()).To(Succeed())
				testFinished <- true
			}()
			done = true
		})

		It("returns the output and true", func() {
			out, finished := br.Next()
			Expect(string(out)).To(Equal("Hello world!"))
			Expect(finished).To(BeTrue())
			<-testFinished
		})
	})

	Context("when the output is being written in chunks", func() {
		var testFinished chan bool

		BeforeEach(func() {
			testFinished = make(chan bool)
			go func() {
				defer GinkgoRecover()

				_, err := writer.Write([]byte("1 "))
				Expect(err).NotTo(HaveOccurred())
				time.Sleep(time.Millisecond * 10)

				_, err = writer.Write([]byte("2 "))
				Expect(err).NotTo(HaveOccurred())
				time.Sleep(time.Millisecond * 10)

				_, err = writer.Write([]byte("3"))
				Expect(err).NotTo(HaveOccurred())

				Expect(writer.Close()).To(Succeed())
				done = true
				testFinished <- true
			}()
			done = false
		})

		It("streams the output", func() {
			var out bytes.Buffer

			var chunk []byte
			var end bool
			for !end {
				chunk, end = br.Next()
				_, err := out.Write(chunk)
				Expect(err).NotTo(HaveOccurred())
			}

			Expect(out.String()).To(Equal("1 2 3"))
			<-testFinished
		})
	})

	Context("when the amount of data written is larger than the buffer size", func() {
		var testFinished chan bool

		BeforeEach(func() {
			br.Buffer = make([]byte, 4)
			testFinished = make(chan bool)

			go func() {
				defer GinkgoRecover()
				_, err := writer.Write([]byte("123456789"))
				Expect(err).NotTo(HaveOccurred())
				Expect(writer.Close()).To(Succeed())
				testFinished <- true
			}()
			done = true
		})

		It("does not consider the file finished until all the data has been streamed", func() {
			var out bytes.Buffer

			var chunk []byte
			var end bool
			for !end {
				chunk, end = br.Next()
				_, err := out.Write(chunk)
				Expect(err).NotTo(HaveOccurred())
			}

			Expect(out.String()).To(Equal("123456789"))
			<-testFinished
		})
	})
})
