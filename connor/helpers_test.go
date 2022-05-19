package connor_test

import (
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/sourcenetwork/defradb/connor"
)

type TestCases map[string][]TestCase

type TestCase struct {
	Name            string
	Conditions      interface{}
	ShouldMatch     bool
	ShouldHaveError bool
}

func (t TestCases) Generate(test func(conditions, data map[string]interface{}) (bool, error)) {
	if test == nil {
		test = Match
	}

	for d, casesl := range t {
		d := d
		cases := casesl

		var data map[string]interface{}
		BeforeEach(func() {
			Expect(json.NewDecoder(strings.NewReader(d)).Decode(&data)).To(Succeed())
		})

		Describe(fmt.Sprintf("with %s as data", d), func() {
			for _, tc := range cases {
				tc := tc

				var conditions map[string]interface{}
				BeforeEach(func() {
					switch c := tc.Conditions.(type) {
					case string:
						Expect(json.NewDecoder(strings.NewReader(c)).Decode(&conditions)).To(Succeed())
					case map[string]interface{}:
						conditions = c
					default:
						Expect(tc.Conditions).To(Or(BeAssignableToTypeOf(string("")), BeAssignableToTypeOf(map[string]interface{}{})))
					}
				})

				var (
					match bool
					err   error
				)
				JustBeforeEach(func() {
					match, err = test(conditions, data)
				})

				Context(fmt.Sprintf("and %s as a condition", tc.Conditions), func() {
					Describe(fmt.Sprintf("should %s", tc.Name), func() {
						if tc.ShouldHaveError {
							It("should return an error", func() {
								Expect(err).ToNot(Succeed())
								Expect(match).To(BeFalse())
							})
						} else {
							It("should not return an error", func() {
								Expect(err).To(Succeed())
							})

							if tc.ShouldMatch {
								It("should match", func() {
									Expect(match).To(BeTrue())
								})
							} else {
								It("shouldn't match", func() {
									Expect(match).To(BeFalse())
								})
							}
						}
					})
				})
			}
		})
	}
}
