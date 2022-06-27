package connor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/sourcenetwork/defradb/connor"
)

var _ = Describe("$nin", func() {
	It("should be registered as an operator", func() {
		Expect(Operators()).To(ContainElement("nin"))
	})

	cases := TestCases{

		`{ "x": 1, "y": 2 }`: []TestCase{
			{
				"not match values which are in the query list",
				`{ "x": { "$nin": [1] } }`,
				false,
				false,
			},
			{
				"match values which are not in the query list",
				`{ "x": { "$nin": [2] } }`,
				true,
				false,
			},
			{
				"not match values which are in the query list with many options",
				`{ "x": { "$nin": [1, 2, 3] } }`,
				false,
				false,
			},
		},
		`{ "a": { "x": 1 }, "y": 2 }`: []TestCase{
			{
				"not match nested object properties",
				`{ "a.x": { "$nin": [1] } }`,
				false,
				false,
			},
			{
				"match nested properties if they are not deep-equal",
				`{ "a": { "$nin": [1] } }`,
				true,
				false,
			},
			{
				"return an error if a query which is not a list is provided",
				`{ "a": { "$nin": 1 } }`,
				false,
				true,
			},
		},
	}

	cases.Generate(nil)
})
