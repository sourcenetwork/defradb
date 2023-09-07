package filter

import (
	"testing"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/planner/mapper"
	"github.com/stretchr/testify/assert"
)

func TestIsComplex(t *testing.T) {
	tests := []struct {
		name        string
		inputFilter map[string]any
		isComplex   bool
	}{
		{
			name: "flat structure",
			inputFilter: map[string]any{
				"name": m("_eq", "John"),
				"age":  m("_gt", 55),
			},
			isComplex: false,
		},
		{
			name: "fields within _and",
			inputFilter: r("_and",
				m("name", m("_eq", "John")),
				m("age", m("_gt", 55)),
			),
			isComplex: false,
		},
		{
			name: "fields within _or and _and",
			inputFilter: r("_and",
				r("_or",
					r("_and",
						m("name", m("_eq", "John")),
						m("age", m("_gt", 30)),
					),
				),
				r("_or",
					m("name", m("_eq", "Islam")),
					m("age", m("_lt", 55)),
				),
			),
			isComplex: false,
		},
		{
			name: "fields within _or and _and",
			inputFilter: r("_or",
				r("_and",
					m("name", m("_eq", "John")),
					m("age", m("_gt", 30)),
				),
				m("verified", m("_eq", true)),
			),
			isComplex: false,
		},
		{
			name: "only 1 relation within _or",
			inputFilter: r("_or",
				m("published", m("rating", m("_gt", 4.0))),
			),
			isComplex: false,
		},
		{
			name: "relation with fields inside _or",
			inputFilter: r("_or",
				m("published", m("rating", m("_gt", 4.0))),
				m("age", m("_gt", 30)),
				m("verified", m("_eq", true)),
			),
			isComplex: true,
		},
		{
			name: "relation not inside _or",
			inputFilter: r("_and",
				r("_or",
					m("age", m("_lt", 30)),
					m("verified", m("_eq", false)),
				),
				r("_or",
					r("_and",
						m("age", m("_gt", 30)),
					),
					m("name", m("_eq", "John")),
				),
				r("_and",
					m("name", m("_eq", "Islam")),
					m("published", m("rating", m("_gt", 4.0))),
				),
			),
			isComplex: false,
		},
		{
			name: "relation with fields inside _and and _or",
			inputFilter: r("_and",
				r("_or",
					m("age", m("_lt", 30)),
					m("verified", m("_eq", false)),
				),
				r("_or",
					r("_and",
						m("published", m("rating", m("_gt", 4.0))),
						m("age", m("_gt", 30)),
					),
					m("name", m("_eq", "John")),
				),
			),
			isComplex: true,
		},
	}

	mapping := getDocMapping()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inputFilter := mapper.ToFilter(request.Filter{Conditions: test.inputFilter}, mapping)
			actual := IsComplex(inputFilter)
			assert.Equal(t, test.isComplex, actual)
		})
	}
}
