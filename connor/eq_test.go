package connor_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/sourcenetwork/defradb/connor"
)

var _ = Describe("$eq", func() {
	It("should be registered as an operator", func() {
		Expect(Operators()).To(ContainElement("eq"))
	})

	Describe("Complex Objects", func() {
		cases := TestCases{
			`{ "x": 1, "y": 2 }`: []TestCase{
				{
					"match a field with the same value",
					`{ "x": { "$eq": 1 } }`,
					true,
					false,
				},
				{
					"not match a field with a different value",
					`{ "x": { "$eq": 2 } }`,
					false,
					false,
				},
				{
					"not match a missing field",
					`{ "a": { "$eq": 1 } }`,
					false,
					false,
				},
				{
					"not match a field with a different value type",
					`{ "x": { "$eq": "1" } }`,
					false,
					false,
				},
			},
			`{ "x": [1] }`: []TestCase{
				{
					"match fields for deep equality",
					`{ "x": [1] }`,
					true,
					false,
				},
			},
			`{ "x": [1, 2, 3] }`: []TestCase{
				{
					"match values which exist within an array",
					`{ "x": 1 }`,
					true,
					false,
				},
			},
			`{ "x": "1", "y": 2 }`: []TestCase{
				{
					"not match a field with a different value type",
					`{ "x": { "$eq": 1 } }`,
					false,
					false,
				},
			},
			`{ "a": { "x": 1 }, "y": 2 }`: []TestCase{
				{
					"match a deep field",
					`{ "a.x": { "$eq": 1 } }`,
					true,
					false,
				},
				{
					"match an object",
					`{ "a": { "x": 1 } }`,
					true,
					false,
				},
				{
					"not match a deep value without a full path to it",
					`{ "a": { "$eq": 1 } }`,
					false,
					false,
				},
			},
			`{ "x": null, "y": 2 }`: []TestCase{
				{
					"match an explicitly null field if null is searched for",
					`{ "x": null }`,
					true,
					false,
				},
				{
					"not match a field which is explicitly null",
					`{ "x": 1 }`,
					false,
					false,
				},
			},
			`{ "x": { "y": 1, "z": 1 } }`: []TestCase{
				{
					"match a deep object explicitly",
					`{ "x": { "$eq": { "y": 1, "z": 1 } } }`,
					true,
					false,
				},
				{
					"not match a deep object explicitly if the values of its fields differ",
					`{ "x": { "$eq": { "y": 2, "z": 2 } } }`,
					false,
					false,
				},
			},
			`{ "x": { "y": [1], "z": 1 } }`: []TestCase{
				{
					"match an object if it has complex properties explicitly searched for",
					`{ "x": { "$eq": { "y": [1] } } }`,
					true,
					false,
				},
				{
					"not match an object if it has complex properties explicitly searched for but values differ",
					`{ "x": { "$eq": { "y": [2] } } }`,
					false,
					false,
				},
			},
			`{ "a": [{ "x": 1 }, { "x": 2 }, { "x": 3 }] }`: []TestCase{
				{
					"match objects which exist within an array",
					`{ "a": { "x": 1 } }`,
					true,
					false,
				},
			},
		}

		cases.Generate(nil)
	})

	Describe("Different Types", func() {
		cases := []struct {
			con  interface{}
			data interface{}

			match  bool
			hasErr bool
		}{
			{
				"test", "test",
				true, false,
			},
			{
				"test", 1,
				false, false,
			},
			{
				int8(10), 10,
				true, false,
			},
			{
				int16(10), 10,
				true, false,
			},
			{
				int32(10), 10,
				true, false,
			},
			{
				int64(10), 10,
				true, false,
			},
			{
				float32(10), 10,
				true, false,
			},
		}

		for _, c := range cases {
			conds := c.con
			data := c.data
			match := c.match
			hasErr := c.hasErr

			Describe(fmt.Sprintf("%T(%v) == %T(%v)", c.con, c.con, c.data, c.data), func() {
				m, err := Match(map[string]interface{}{
					"x": map[string]interface{}{"$eq": conds},
				}, map[string]interface{}{
					"x": data,
				})

				if hasErr {
					It("should return an error", func() {
						Expect(err).ToNot(Succeed())
					})
				} else {
					It("should not return an error", func() {
						Expect(err).To(Succeed())
					})
				}

				if match {
					It("should match", func() {
						Expect(m).To(BeTrue())
					})
				} else {
					It("should not match", func() {
						Expect(m).To(BeFalse())

					})
				}
			})
		}
	})
})
