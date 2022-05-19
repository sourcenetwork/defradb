package fields_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFields(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fields Suite")
}
