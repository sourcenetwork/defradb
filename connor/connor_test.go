package connor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/sourcenetwork/defradb/connor"
)

var _ = Describe("Connor", func() {
	Describe("with a malformed operator", func() {
		_, err := MatchWith("malformed", nil, nil)

		It("should return an error", func() {
			Expect(err).ToNot(BeNil())
		})

		It("should provide a descriptive error", func() {
			Expect(err.Error()).To(Equal("operator should have '$' prefix"))
		})

		It("should return a short error", func() {
			Expect(len(err.Error()) < 80).To(BeTrue(), "error message should be less than 80 characters long")
		})
	})

	Describe("with an invalid/unknown operator", func() {
		_, err := MatchWith("$invalid", nil, nil)

		It("should return an error", func() {
			Expect(err).ToNot(BeNil())
		})

		It("should provide a descriptive error", func() {
			Expect(err.Error()).To(Equal("unknown operator 'invalid'"))
		})

		It("should return a short error", func() {
			Expect(len(err.Error()) < 80).To(BeTrue(), "error message should be less than 80 characters long")
		})
	})
})
