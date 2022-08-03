package fixtures

import "fmt"

type ratio [2]float64

func newRatio(i, j float64) (ratio, error) {
	if i == 0 || j == 0 {
		return ratio{}, fmt.Errorf("ratio can't contain zero")
	}
	return ratio{i, j}, nil
}

func (r ratio) Decimal() float64 {
	return r[0] / r[1]
}

// Equal r == r2
func (r ratio) Equal(r2 ratio) bool {
	return r.Decimal() == r2.Decimal()
}

// Less r < r2
func (r ratio) Less(r2 ratio) bool {
	return r.Decimal() < r2.Decimal()
}

// Greater r > r2
func (r ratio) Greater(r2 ratio) bool {
	return r.Decimal() > r2.Decimal()
}

// A returns A side of the A:B ratio
func (r ratio) A() float64 { return r[0] }

// B returns B side of the A:B ratio
func (r ratio) B() float64 { return r[1] }

func (r ratio) Invert() ratio {
	r2, _ := newRatio(r[1], r[0])
	return r2
}

// Noramlize will return a new ratio with both
// sides of the ratio mulitplied by a constant.
//
// eg: ratio(1:1).Normalize(8) => ratio(8:8)
func (r ratio) Normalize(c float64) ratio {
	r2, _ := newRatio(r[0]*c, r[1]*c)
	return r2
}
