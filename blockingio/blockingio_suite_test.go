package blockingio_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBlockingio(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Blockingio Suite")
}
