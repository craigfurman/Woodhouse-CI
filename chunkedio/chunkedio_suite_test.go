package chunkedio_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func Testchunkedio(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "chunkedio Suite")
}
