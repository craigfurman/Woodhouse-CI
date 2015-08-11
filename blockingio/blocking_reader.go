package blockingio

import (
	"io"
	"log"
	"os"
	"time"
)

var buf = make([]byte, 1024)

type BlockingReader struct {
	Output      *os.File
	DoneWriting func() bool
}

func (r *BlockingReader) Next() ([]byte, bool) {
	var err error
	bytesRead := 0
	done := false
	for bytesRead < 1 && !done {
		done = r.DoneWriting()
		bytesRead, err = r.Output.Read(buf)
		if err != nil && err != io.EOF {
			log.Printf("error streaming output file. Cause: %v\n", err)
		}
		if !done {
			time.Sleep(time.Millisecond * 25)
		}
	}
	return buf[:bytesRead], done
}

func (r *BlockingReader) Close() error {
	return r.Output.Close()
}
