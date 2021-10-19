// Copyright 2020 Source Inc.
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

	testUtils "github.com/sourcenetwork/defradb/db/tests"
)

func TestQuerySimpleWithDocKeyFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with basic filter (key by DocKey arg)",
		Query: `query {
					users(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 21
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"Name": "John",
				"Age":  uint64(21),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithKeyFilterBlock(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with basic filter (key by filter block)",
		Query: `query {
					users(filter: {_key: {_eq: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"}}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 21
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"Name": "John",
				"Age":  uint64(21),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithStringFilterBlock(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with basic filter (Name)",
		Query: `query {
					users(filter: {Name: {_eq: "John"}}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 21
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"Name": "John",
				"Age":  uint64(21),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithStringFilterBlockAndSelect(t *testing.T) {
	tests := []testUtils.QueryTestCase{
		{
			Description: "Simple query with basic filter and selection",
			Query: `query {
						users(filter: {Name: {_eq: "John"}}) {
							Name
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			Results: []map[string]interface{}{
				{
					"Name": "John",
				},
			},
		},
		{
			Description: "Simple query with basic filter and selection (diff from filter)",
			Query: `query {
						users(filter: {Name: {_eq: "John"}}) {
							Age
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			Results: []map[string]interface{}{
				{
					"Age": uint64(21),
				},
			},
		},
		{
			Description: "Simple query with basic filter(name), no results",
			Query: `query {
						users(filter: {Name: {_eq: "Bob"}}) {
							Name
							Age
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			Results: []map[string]interface{}{},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQuerySimpleWithNumberEqualsFilterBlock(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with basic filter(age)",
		Query: `query {
					users(filter: {Age: {_eq: 21}}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 21
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"Name": "John",
				"Age":  uint64(21),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithNumberGreaterThanFilterBlock(t *testing.T) {
	tests := []testUtils.QueryTestCase{
		{
			Description: "Simple query with basic filter(age), greater than",
			Query: `query {
						users(filter: {Age: {_gt: 20}}) {
							Name
							Age
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			Results: []map[string]interface{}{
				{
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
		{
			Description: "Simple query with basic filter(age), no results",
			Query: `query {
						users(filter: {Age: {_gt: 40}}) {
							Name
							Age
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 32
				}`)},
			},
			Results: []map[string]interface{}{},
		},
		{
			Description: "Simple query with basic filter(age), multiple results",
			Query: `query {
						users(filter: {Age: {_gt: 20}}) {
							Name
							Age
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 32
				}`)},
			},
			Results: []map[string]interface{}{
				{
					"Name": "Bob",
					"Age":  uint64(32),
				},
				{
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQuerySimpleWithNumberGreaterThanAndNumberLessThanFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with logical compound filter (and)",
		Query: `query {
					users(filter: {_and: [{Age: {_gt: 20}}, {Age: {_lt: 50}}]}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 21
			}`),
				(`{
				"Name": "Bob",
				"Age": 32
			}`),
				(`{
				"Name": "Carlo",
				"Age": 55
			}`),
				(`{
				"Name": "Alice",
				"Age": 19
			}`)},
		},
		Results: []map[string]interface{}{
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

func TestQuerySimpleWithNumberEqualToXOrYFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with logical compound filter (or)",
		Query: `query {
					users(filter: {_or: [{Age: {_eq: 55}}, {Age: {_eq: 19}}]}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 21
			}`),
				(`{
				"Name": "Bob",
				"Age": 32
			}`),
				(`{
				"Name": "Carlo",
				"Age": 55
			}`),
				(`{
				"Name": "Alice",
				"Age": 19
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"Name": "Alice",
				"Age":  uint64(19),
			},
			{
				"Name": "Carlo",
				"Age":  uint64(55),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithNumberInFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with special filter (or)",
		Query: `query {
					users(filter: {Age: {_in: [19, 40, 55]}}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 21
			}`),
				(`{
				"Name": "Bob",
				"Age": 32
			}`),
				(`{
				"Name": "Carlo",
				"Age": 55
			}`),
				(`{
				"Name": "Alice",
				"Age": 19
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"Name": "Alice",
				"Age":  uint64(19),
			},
			{
				"Name": "Carlo",
				"Age":  uint64(55),
			},
		},
	}

	executeTestCase(t, test)
}
