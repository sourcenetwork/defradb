package connor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/sourcenetwork/defradb/connor"
)

var _ = Describe("$ne", func() {
	It("should be registered as an operator", func() {
		Expect(Operators()).To(ContainElement("ne"))
	})

	cases := TestCases{
		`{ "x": 1, "y": 2 }`: []TestCase{
			{
				"not match when the values are equal",
				`{ "x": { "$ne": 1 } }`,
				false,
				false,
			},
			{
				"match when the values are different",
				`{ "x": { "$ne": 2 } }`,
				true,
				false,
			},
			{
				"match when the field is not present",
				`{ "a": { "$ne": 1 } }`,
				true,
				false,
			},
			{
				"match when the types are different",
				`{ "x": { "$ne": "1" } }`,
				true,
				false,
			},
		},
		`{ "a": { "x": 1 }, "y": 2 }`: []TestCase{
			{
				"not match when a nested property has the same value",
				`{ "a.x": { "$ne": 1 } }`,
				false,
				false,
			},
			{
				"match when a nested property has a different value",
				`{ "a": { "$ne": 1 } }`,
				true,
				false,
			},
			{
				"match when a missing nested property is tested for value equality",
				`{ "a": { "$ne": { "z": 1 } } }`,
				true,
				false,
			},
		},
	}

	cases.Generate(nil)
})
