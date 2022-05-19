package connor_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/sourcenetwork/defradb/connor"
)

var _ = Describe("$gt", func() {
	now := time.Now()

	It("should be registered as an operator", func() {
		Expect(Operators()).To(ContainElement("gt"))
	})

	Describe("Basic Cases", func() {
		cases := TestCases{
			`{ "x": 1, "y": 2 }`: []TestCase{
				{
					"match numbers which are greater",
					`{ "x": { "$gt": 0 } }`,
					true,
					false,
				},
				{
					"not match numbers which are equal",
					`{ "x": { "$gt": 1 } }`,
					false,
					false,
				},
				{
					"not match numbers which are less",
					`{ "x": { "$gt": 2 } }`,
					false,
					false,
				},
				{
					"match numbers by up-casting them as necessary",
					`{ "x": { "$gt": 0.5 } }`,
					true,
					false,
				},
			},
			`{ "a": { "x": 1 }, "y": 2 }`: []TestCase{
				{
					"match nested object properties",
					`{ "a.x": { "$gt": 0 } }`,
					true,
					false,
				},
				{
					"not match nested object properties which are less",
					`{ "a": { "$gt": 2 } }`,
					false,
					false,
				},
			},
			`{ "x": "5", "y": 2 }`: []TestCase{
				{
					"not match strings logically when they are lexicographically less",
					`{ "x": { "$gt": "7" } }`,
					false,
					false,
				},
				{
					"not match strings logically when they are lexicographically equal",
					`{ "x": { "$gt": "5" } }`,
					false,
					false,
				},
				{
					"not match across different value types",
					`{ "x": { "$gt": 0 } }`,
					false,
					false,
				},
			},
			`{ "x": "b", "y": 2 }`: []TestCase{
				{
					"match strings which are lexicographically larger",
					`{ "x": { "$gt": "a" } }`,
					true,
					false,
				},
				{
					"not match strings which are lexicographically smaller",
					`{ "x": { "$gt": "c" } }`,
					false,
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
				"abc", "def",
				true, false,
			},
			{
				"abc", "abc",
				false, false,
			},
			{
				"abc", "aaa",
				false, false,
			},
			{
				"test", 1,
				false, false,
			},
			{
				int8(10), 100,
				true, false,
			},
			{
				int16(10), 106,
				true, false,
			},
			{
				int32(10), 15,
				true, false,
			},
			{
				int64(10), 12,
				true, false,
			},
			{
				float32(10), 11,
				true, false,
			},
			{
				int64(10), float32(10),
				false, false,
			},
			{
				int64(10), "test",
				false, false,
			},
			{
				now, now,
				false, false,
			},
			{
				now, now.Add(time.Second),
				true, false,
			},
			{
				now, now.Add(-time.Second),
				false, false,
			},
			{
				now, 10,
				false, false,
			},
			{
				[]int{10}, []int{12},
				false, true,
			},
		}

		for _, c := range cases {
			conds := c.con
			data := c.data
			match := c.match
			hasErr := c.hasErr

			Describe(fmt.Sprintf("%T(%v) == %T(%v)", c.con, c.con, c.data, c.data), func() {
				m, err := Match(map[string]interface{}{
					"x": map[string]interface{}{"$gt": conds},
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
