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

func TestQuerySimple_WithCountWithGroupBy_OnEmptyCollection_ReturnsZero(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.Request{
				Request: `query {
					COUNT(Users: {groupBy: [Age]})
				}`,
				Results: map[string]any{
					"COUNT": 0,
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithCountWithGroupBy_OnSingleField_CountsDistinctValues(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					COUNT(Users: {groupBy: [Age]})
				}`,
				Results: map[string]any{
					"COUNT": 2,
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithCountWithGroupBy_OnMultipleFields_CountsDistinctCombinations(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					COUNT(Users: {groupBy: [Age, Name]})
				}`,
				Results: map[string]any{
					"COUNT": 3,
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithCountWithGroupBy_AllSameGroupValue_ReturnsOne(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 32
				}`,
			},
			&action.Request{
				Request: `query {
					COUNT(Users: {groupBy: [Age]})
				}`,
				Results: map[string]any{
					"COUNT": 1,
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithCountWithGroupByAndFilter_CountsDistinctValuesAfterFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Chris",
					"Age": 25
				}`,
			},
			&action.Request{
				Request: `query {
					COUNT(Users: {groupBy: [Age], filter: {Age: {_gt: 20}}})
				}`,
				Results: map[string]any{
					"COUNT": 2,
				},
			},
		},
	}

	executeTestCase(t, test)
}
