package chunkedio

import (
	"io"
	"log"
	"time"
)

type ChunkedReader struct {
	Output      io.ReadCloser
	DoneWriting func() bool
	Buffer      []byte
}

func (r *ChunkedReader) Next() ([]byte, bool) {
	var err error
	bytesRead := 0
	done := false
	for bytesRead < 1 && !done {
		done = r.DoneWriting()
		bytesRead, err = r.Output.Read(r.Buffer)
		if err != nil && err != io.EOF {
			log.Printf("error streaming output file. Cause: %v\n", err)
		}
		if !done {
			time.Sleep(time.Millisecond * 25)
		}
		if bytesRead == len(r.Buffer) {
			done = false
		}
	}
	return r.Buffer[:bytesRead], done
}

func (r *ChunkedReader) Close() error {
	return r.Output.Close()
}
