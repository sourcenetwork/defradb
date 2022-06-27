package connor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/sourcenetwork/defradb/connor"
)

var _ = Describe("$in", func() {
	It("should be registered as an operator", func() {
		Expect(Operators()).To(ContainElement("in"))
	})

	cases := TestCases{

		`{ "x": 1, "y": 2 }`: []TestCase{
			{
				"match values which are in the query list",
				`{ "x": { "$in": [1] } }`,
				true,
				false,
			},
			{
				"not match values which are not in the query list",
				`{ "x": { "$in": [2] } }`,
				false,
				false,
			},
			{
				"match values which are in the query list with many options",
				`{ "x": { "$in": [1, 2, 3] } }`,
				true,
				false,
			},
		},
		`{ "a": { "x": 1 }, "y": 2 }`: []TestCase{
			{
				"match nested object properties",
				`{ "a.x": { "$in": [1] } }`,
				true,
				false,
			},
			{
				"not match nested properties if their full key path is not provided",
				`{ "a": { "$in": [1] } }`,
				false,
				false,
			},
			{
				"return an error if a query which is not a list is provided",
				`{ "a": { "$in": 1 } }`,
				false,
				true,
			},
		},
	}

	cases.Generate(nil)
})
