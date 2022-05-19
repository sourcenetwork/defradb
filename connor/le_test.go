package connor_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/sourcenetwork/defradb/connor"
)

var _ = Describe("$le", func() {
	now := time.Now()

	It("should be registered as an operator", func() {
		Expect(Operators()).To(ContainElement("le"))
	})

	Describe("Basic Cases", func() {
		cases := TestCases{
			`{ "x": 1, "y": 2 }`: []TestCase{
				{
					"not match numbers which are greater",
					`{ "x": { "$le": 0 } }`,
					false,
					false,
				},
				{
					"match numbers which are equal",
					`{ "x": { "$le": 1 } }`,
					true,
					false,
				},
				{
					"match numbers which are less",
					`{ "x": { "$le": 2 } }`,
					true,
					false,
				},
				{
					"match numbers by up-casting them as necessary",
					`{ "x": { "$le": 1.3 } }`,
					true,
					false,
				},
			},
			`{ "a": { "x": 1 }, "y": 2 }`: []TestCase{
				{
					"match nested object properties",
					`{ "a.x": { "$le": 2 } }`,
					true,
					false,
				},
				{
					"not match nested object properties which are less",
					`{ "a": { "$le": 0 } }`,
					false,
					false,
				},
			},
			`{ "x": "5", "y": 2 }`: []TestCase{
				{
					"not match strings logically when they are lexicographically less",
					`{ "x": { "$le": "3" } }`,
					false,
					false,
				},
				{
					"not match across different value types",
					`{ "x": { "$le": 10 } }`,
					false,
					false,
				},
			},
			`{ "x": "b", "y": 2 }`: []TestCase{
				{
					"match strings which are lexicographically larger",
					`{ "x": { "$le": "c" } }`,
					true,
					false,
				},
				{
					"not match strings which are lexicographically smaller",
					`{ "x": { "$le": "a" } }`,
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
				false, false,
			},
			{
				"abc", "abc",
				true, false,
			},
			{
				"abc", "aaa",
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
				int64(10), 12,
				false, false,
			},
			{
				float32(10), 9,
				true, false,
			},
			{
				int64(10), float32(10),
				true, false,
			},
			{
				int64(10), "test",
				false, false,
			},
			{
				now, now,
				true, false,
			},
			{
				now, now.Add(time.Second),
				false, false,
			},
			{
				now, now.Add(-time.Second),
				true, false,
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
					"x": map[string]interface{}{"$le": conds},
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
