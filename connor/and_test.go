package connor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/sourcenetwork/defradb/connor"
)

var _ = Describe("$and", func() {
	It("should be registered as an operator", func() {
		Expect(Operators()).To(ContainElement("and"))
	})

	cases := TestCases{
		`{ "x": 1, "y": 2 }`: []TestCase{
			{"match with a single value", `{ "x": { "$and": [1] } }`, true, false},
			{"not match with a single value", `{ "x": { "$and": [2] } }`, false, false},
			{"match with a single operation", `{ "x": { "$and": [{ "$eq": 1 }] } }`, true, false},
			{"not match with a single operation", `{ "x": { "$and": [{ "$eq": 2 }] } }`, false, false},
			{"not match with multiple values", `{ "x": { "$and": [1, 2] } }`, false, false},
			{"error without an array of values", `{ "x": { "$and": 1 } }`, false, true},
		},
		`{ "a": { "x": 1 }, "y": 2 }`: []TestCase{
			{"not match when a nested operator doesn't match", `{ "x": { "$and": [{ "$eq": 1 }] } }`, false, false},
			{"match when nested operators all match", `{ "a.x": { "$and": [{ "$in": [1, 3] }, { "$in": [1, 2] }] } }`, true, false},
		},
	}

	cases.Generate(nil)
})
