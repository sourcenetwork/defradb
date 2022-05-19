package connor_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConnor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Connor Suite")
}
