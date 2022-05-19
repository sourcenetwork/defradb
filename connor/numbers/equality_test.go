package numbers_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/sourcenetwork/defradb/connor/numbers"
)

var _ = Describe("Equality", func() {
	Describe("Equal", func() {
		cases := []struct {
			a interface{}
			b interface{}

			matches bool
		}{
			{int64(10), int64(10), true},
			{int64(10), int64(0), false},

			{float64(10), int64(10), true},
			{int64(10), float64(10), true},
			{float64(10), int64(12), false},
			{int64(10), float64(12), false},

			{int8(5), float64(5), true},

			{"test", int64(5), false},
			{int64(5), "test", false},
			{"test", float64(5), false},
			{float64(5), "test", false},
			{"test", "test", false},
		}

		for _, c := range cases {
			a := c.a
			b := c.b

			if c.matches {
				It(fmt.Sprintf("should determine that %T(%v) == %T(%v)", a, a, b, b), func() {
					Expect(numbers.Equal(a, b)).To(BeTrue())
				})
			} else {
				It(fmt.Sprintf("should determine that %T(%v) != %T(%v)", a, a, b, b), func() {
					Expect(numbers.Equal(a, b)).To(BeFalse())
				})
			}
		}
	})
})
