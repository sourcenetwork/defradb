// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimple_WithNotEqualToXFilter_NoError(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with logical compound filter (not)",
		Request: `query {
					Users(filter: {_not: {Age: {_eq: 55}}}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
				`{
					"Name": "Bob",
					"Age": 32
				}`,
				`{
					"Name": "Carlo",
					"Age": 55
				}`,
				`{
					"Name": "Alice",
					"Age": 19
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Bob",
				"Age":  uint64(32),
			},
			{
				"Name": "Alice",
				"Age":  uint64(19),
			},
			{
				"Name": "John",
				"Age":  uint64(21),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithNotAndComparisonXFilter_NoError(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with _not filter with _gt condition)",
		Request: `query {
					Users(filter: {_not: {Age: {_gt: 20}}}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
				`{
					"Name": "Bob",
					"Age": 32
				}`,
				`{
					"Name": "Carlo",
					"Age": 55
				}`,
				`{
					"Name": "Alice",
					"Age": 19
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Alice",
				"Age":  uint64(19),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithNotEqualToXorYFilter_NoError(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with logical compound filter (not)",
		Request: `query {
					Users(filter: {_not: {_or: [{Age: {_eq: 55}}, {Name: {_eq: "Alice"}}]}}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
				`{
					"Name": "Bob",
					"Age": 32
				}`,
				`{
					"Name": "Carlo",
					"Age": 55
				}`,
				`{
					"Name": "Alice",
					"Age": 19
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Bob",
				"Age":  uint64(32),
			},
			{
				"Name": "John",
				"Age":  uint64(21),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithEmptyNotFilter_ReturnError(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with empty logical compound filter (not) returns empty result set",
		Request: `query {
					Users(filter: {_not: {}}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
				`{
					"Name": "Bob",
					"Age": 32
				}`,
				`{
					"Name": "Carlo",
					"Age": 55
				}`,
				`{
					"Name": "Alice",
					"Age": 19
				}`,
			},
		},
		Results: []map[string]any{},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithNotEqualToXAndNotYFilter_NoError(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with logical compound filter (not)",
		Request: `query {
					Users(filter: {_not: {Age: {_eq: 55}, _not: {Name: {_eq: "Carlo"}}}}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
				`{
					"Name": "Bob",
					"Age": 32
				}`,
				`{
					"Name": "Carlo",
					"Age": 55
				}`,
				`{
					"Name": "Alice",
					"Age": 19
				}`,
				`{
					"Name": "Frank",
					"Age": 55
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Bob",
				"Age":  uint64(32),
			},
			{
				"Name": "Alice",
				"Age":  uint64(19),
			},
			{
				"Name": "John",
				"Age":  uint64(21),
			},
			{
				"Name": "Carlo",
				"Age":  uint64(55),
			},
		},
	}

	executeTestCase(t, test)
}
