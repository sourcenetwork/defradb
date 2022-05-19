package numbers_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/sourcenetwork/defradb/connor/numbers"
)

var _ = Describe("Upcast", func() {
	Describe("TryUpcast", func() {
		cases := []struct {
			in  interface{}
			out interface{}
		}{
			{int8(10), int64(10)},
			{int16(10), int64(10)},
			{int32(10), int64(10)},
			{int64(10), int64(10)},
			{float32(10), float64(10)},
			{float64(10), float64(10)},
			{"test", "test"},
		}

		for _, c := range cases {
			in := c.in
			out := c.out

			It(fmt.Sprintf("should convert %T(%v) to %T(%v)", in, in, out, out), func() {
				Expect(numbers.TryUpcast(in)).To(Equal(out))
			})
		}
	})
})
