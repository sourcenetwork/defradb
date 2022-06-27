package connor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/sourcenetwork/defradb/connor"
)

var _ = Describe("$or", func() {
	It("should be registered as an operator", func() {
		Expect(Operators()).To(ContainElement("or"))
	})

	cases := TestCases{
		`{ "x": 1, "y": 2 }`: []TestCase{
			{
				"match equality of values implicitly",
				`{ "x": { "$or": [1] } }`,
				true,
				false,
			},
			{
				"not match inequality of values implicitly",
				`{ "x": { "$or": [2] } }`,
				false,
				false,
			},
			{
				"match using explicit value comparison operators",
				`{ "x": { "$or": [{ "$eq": 1 }] } }`,
				true,
				false,
			},
			{
				"return an error if you do not provide a list of options",
				`{ "x": { "$or": 2 } }`,
				false,
				true,
			},
		},

		`{ "a": { "x": 1 }, "y": 2 }`: []TestCase{
			{
				"not match if values are not deep-equal with an explicit operator",
				`{ "x": { "$or": [{ "$eq": 1 }] } }`,
				false,
				false,
			},
			{
				"match if a complex value comparison is performed",
				`{ "a": { "$or": [{ "x": { "$in": [1] } }] } }`,
				true,
				false,
			},
		},
	}

	cases.Generate(nil)

})
