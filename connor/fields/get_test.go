package fields_test

import (
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/sourcenetwork/defradb/connor/fields"
)

var _ = Describe("Get", func() {
	type TestCase struct {
		field string
		value string
		found bool
	}

	cases := map[string][]TestCase{
		`{"x": 1}`: {
			{
				"x",
				"1",
				true,
			},
			{
				"y",
				"null",
				false,
			},
			{
				"x.y",
				`null`,
				false,
			},
		},
		`{"x": null}`: {
			{
				"x",
				"null",
				true,
			},
		},
		`{"x": { "y": 1 }}`: {
			{
				"x.y",
				"1",
				true,
			},
			{
				"x",
				`{ "y": 1 }`,
				true,
			},
		},
		`{"x": [ { "y": 1 }, { "y" : 2 } ]}`: {
			{
				"x.0.y",
				`1`,
				true,
			},
			{
				"x.1.y",
				`2`,
				true,
			},
		},
		`{"x": [ { "y": [ 5,6] }]}`: {
			{
				"x.0.y.0",
				`5`,
				true,
			},
			{
				"x.0.y.1",
				`6`,
				true,
			},
			{
				"x.0.y.2",
				`null`,
				false,
			},
			{
				"x.-1.y.2",
				`null`,
				false,
			},
		},
		`{"x": [ { "y": 1 } ]}`: {
			{
				"x.0.z",
				`null`,
				false,
			},
			{
				"x.3.z",
				`null`,
				false,
			},
		},
	}

	for dataStr, cs := range cases {
		dataStr := dataStr
		cs := cs
		Describe(fmt.Sprintf("with %s as data", dataStr), func() {
			for _, c := range cs {
				c := c
				Context(fmt.Sprintf("getting the field %s", c.field), func() {
					var (
						data     map[string]interface{}
						value    interface{}
						expected interface{}
						found    bool
					)

					BeforeEach(func() {
						Expect(json.NewDecoder(strings.NewReader(dataStr)).Decode(&data)).To(Succeed())
					})
					BeforeEach(func() {
						Expect(json.NewDecoder(strings.NewReader(c.value)).Decode(&expected)).To(Succeed())
					})

					Context("Get()", func() {
						JustBeforeEach(func() {
							value, found = fields.Get(data, c.field)
						})

						if c.found {
							Specify("should find the field", func() {
								Expect(found).To(BeTrue())
							})
						} else {
							Specify("should not find the field", func() {
								Expect(found).To(BeFalse())
							})
						}

						Specify(fmt.Sprintf("should return %s", c.value), func() {
							if expected == nil {
								Expect(value).To(BeNil())
							} else {
								Expect(value).To(BeEquivalentTo(expected))
							}
						})
					})

					Context("TryGet()", func() {
						JustBeforeEach(func() {
							value = fields.TryGet(data, c.field)
						})

						Specify(fmt.Sprintf("should return %s", c.value), func() {
							if expected == nil {
								Expect(value).To(BeNil())
							} else {
								Expect(value).To(BeEquivalentTo(expected))
							}
						})
					})
				})
			}
		})
	}
})
