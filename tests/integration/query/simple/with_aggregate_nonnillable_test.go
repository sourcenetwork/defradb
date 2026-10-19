// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// Int!

func TestQuerySimpleWithSumOnNonNillableInt(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { score: Int! }`,
			},
			&action.AddDoc{Doc: `{"score": 10}`},
			&action.AddDoc{Doc: `{"score": 20}`},
			&action.AddDoc{Doc: `{"score": 30}`},
			&action.Request{
				Request: `query { SUM(Users: {field: score}) }`,
				Results: map[string]any{
					"SUM": int64(60),
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithAverageOnNonNillableInt(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { score: Int! }`,
			},
			&action.AddDoc{Doc: `{"score": 10}`},
			&action.AddDoc{Doc: `{"score": 20}`},
			&action.AddDoc{Doc: `{"score": 30}`},
			&action.Request{
				Request: `query { AVG(Users: {field: score}) }`,
				Results: map[string]any{
					"AVG": float64(20),
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithMaxOnNonNillableInt(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { score: Int! }`,
			},
			&action.AddDoc{Doc: `{"score": 10}`},
			&action.AddDoc{Doc: `{"score": 20}`},
			&action.AddDoc{Doc: `{"score": 30}`},
			&action.Request{
				Request: `query { MAX(Users: {field: score}) }`,
				Results: map[string]any{
					"MAX": float64(30),
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithMinOnNonNillableInt(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { score: Int! }`,
			},
			&action.AddDoc{Doc: `{"score": 10}`},
			&action.AddDoc{Doc: `{"score": 20}`},
			&action.AddDoc{Doc: `{"score": 30}`},
			&action.Request{
				Request: `query { MIN(Users: {field: score}) }`,
				Results: map[string]any{
					"MIN": float64(10),
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithCountOnNonNillableInt(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { score: Int! }`,
			},
			&action.AddDoc{Doc: `{"score": 10}`},
			&action.AddDoc{Doc: `{"score": 20}`},
			&action.AddDoc{Doc: `{"score": 30}`},
			&action.Request{
				Request: `query { COUNT(Users: {}) }`,
				Results: map[string]any{
					"COUNT": int64(3),
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// Float64!

func TestQuerySimpleWithSumOnNonNillableFloat64(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { rating: Float64! }`,
			},
			&action.AddDoc{Doc: `{"rating": 1.5}`},
			&action.AddDoc{Doc: `{"rating": 2.5}`},
			&action.Request{
				Request: `query { SUM(Users: {field: rating}) }`,
				Results: map[string]any{
					"SUM": float64(4),
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithAverageOnNonNillableFloat64(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { rating: Float64! }`,
			},
			&action.AddDoc{Doc: `{"rating": 1.5}`},
			&action.AddDoc{Doc: `{"rating": 2.5}`},
			&action.Request{
				Request: `query { AVG(Users: {field: rating}) }`,
				Results: map[string]any{
					"AVG": float64(2),
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithMaxOnNonNillableFloat64(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { rating: Float64! }`,
			},
			&action.AddDoc{Doc: `{"rating": 1.5}`},
			&action.AddDoc{Doc: `{"rating": 3.5}`},
			&action.Request{
				Request: `query { MAX(Users: {field: rating}) }`,
				Results: map[string]any{
					"MAX": float64(3.5),
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithMinOnNonNillableFloat64(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { rating: Float64! }`,
			},
			&action.AddDoc{Doc: `{"rating": 1.5}`},
			&action.AddDoc{Doc: `{"rating": 3.5}`},
			&action.Request{
				Request: `query { MIN(Users: {field: rating}) }`,
				Results: map[string]any{
					"MIN": float64(1.5),
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithCountOnNonNillableFloat64(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { rating: Float64! }`,
			},
			&action.AddDoc{Doc: `{"rating": 1.5}`},
			&action.AddDoc{Doc: `{"rating": 2.5}`},
			&action.Request{
				Request: `query { COUNT(Users: {}) }`,
				Results: map[string]any{
					"COUNT": int64(2),
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// Float32!

func TestQuerySimpleWithSumOnNonNillableFloat32(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { weight: Float32! }`,
			},
			&action.AddDoc{Doc: `{"weight": 60.0}`},
			&action.AddDoc{Doc: `{"weight": 80.0}`},
			&action.Request{
				Request: `query { SUM(Users: {field: weight}) }`,
				Results: map[string]any{
					"SUM": float64(140),
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithAverageOnNonNillableFloat32(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { weight: Float32! }`,
			},
			&action.AddDoc{Doc: `{"weight": 60.0}`},
			&action.AddDoc{Doc: `{"weight": 80.0}`},
			&action.Request{
				Request: `query { AVG(Users: {field: weight}) }`,
				Results: map[string]any{
					"AVG": float64(70),
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithMaxOnNonNillableFloat32(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { weight: Float32! }`,
			},
			&action.AddDoc{Doc: `{"weight": 60.0}`},
			&action.AddDoc{Doc: `{"weight": 80.0}`},
			&action.Request{
				Request: `query { MAX(Users: {field: weight}) }`,
				Results: map[string]any{
					"MAX": float64(80),
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithMinOnNonNillableFloat32(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { weight: Float32! }`,
			},
			&action.AddDoc{Doc: `{"weight": 60.0}`},
			&action.AddDoc{Doc: `{"weight": 80.0}`},
			&action.Request{
				Request: `query { MIN(Users: {field: weight}) }`,
				Results: map[string]any{
					"MIN": float64(60),
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithCountOnNonNillableFloat32(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { weight: Float32! }`,
			},
			&action.AddDoc{Doc: `{"weight": 60.0}`},
			&action.AddDoc{Doc: `{"weight": 80.0}`},
			&action.Request{
				Request: `query { COUNT(Users: {}) }`,
				Results: map[string]any{
					"COUNT": int64(2),
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
