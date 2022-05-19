package connor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/sourcenetwork/defradb/connor"
)

var _ = Describe("$contains", func() {
	It("should be registered as an operator", func() {
		Expect(Operators()).To(ContainElement("contains"))
	})

	cases := TestCases{
		`{"x":1}`: []TestCase{
			{
				"error if a non-string value is provided",
				`{"x":{"$contains":"abc"}}`,
				false,
				true,
			},
		},
		`{"x":"abc"}`: []TestCase{
			{
				"match a complete string",
				`{"x":{"$contains":"abc"}}`,
				true,
				false,
			},
			{
				"match a partial suffix",
				`{"x":{"$contains":"bc"}}`,
				true,
				false,
			},
			{
				"match a partial prefix",
				`{"x":{"$contains":"ab"}}`,
				true,
				false,
			},
			{
				"not match a different string",
				`{"x":{"$contains":"xyz"}}`,
				false,
				false,
			},
			{
				"not match a missing field",
				`{"y":{"$contains":"xyz"}}`,
				false,
				false,
			},
		},
	}

	cases.Generate(nil)
})
