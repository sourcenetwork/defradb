// Copyright 2022 Democratized Data Foundation
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

func TestQuerySimpleWithDocKeyFilter(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple query with basic filter (key by DocKey arg)",
			Request: `query {
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
		},
		{
			Description: "Simple query with basic filter (key by DocKey arg), no results",
			Request: `query {
						users(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009g") {
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
		{
			Description: "Simple query with basic filter (key by DocKey arg), partial results",
			Request: `query {
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
					}`),
					(`{
						"Name": "Bob",
						"Age": 32
					}`)},
			},
			Results: []map[string]interface{}{
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

func TestQuerySimpleWithDocKeysFilter(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple query with basic filter (single key by DocKeys arg)",
			Request: `query {
						users(dockeys: ["bae-52b9170d-b77a-5887-b877-cbdbb99b009f"]) {
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
			Description: "Simple query with basic filter (single key by DocKeys arg), no results",
			Request: `query {
						users(dockeys: ["bae-52b9170d-b77a-5887-b877-cbdbb99b009g"]) {
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
		{
			Description: "Simple query with basic filter (duplicate key by DocKeys arg), partial results",
			Request: `query {
						users(dockeys: ["bae-52b9170d-b77a-5887-b877-cbdbb99b009f", "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"]) {
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
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
		{
			Description: "Simple query with basic filter (multiple key by DocKeys arg), partial results",
			Request: `query {
						users(dockeys: ["bae-52b9170d-b77a-5887-b877-cbdbb99b009f", "bae-1378ab62-e064-5af4-9ea6-49941c8d8f94"]) {
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
						"Name": "Jim",
						"Age": 27
					}`)},
			},
			Results: []map[string]interface{}{
				{
					"Name": "Jim",
					"Age":  uint64(27),
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

func TestQuerySimpleWithKeyFilterBlock(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic filter (key by filter block)",
		Request: `query {
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
				}`),
				(`{
				"Name": "Bob",
				"Age": 32
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
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic filter (Name)",
		Request: `query {
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
				}`),
				(`{
				"Name": "Bob",
				"Age": 32
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
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple query with basic filter and selection",
			Request: `query {
						users(filter: {Name: {_eq: "John"}}) {
							Name
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
					"Name": "John",
				},
			},
		},
		{
			Description: "Simple query with basic filter and selection (diff from filter)",
			Request: `query {
						users(filter: {Name: {_eq: "John"}}) {
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
					"Age": uint64(21),
				},
			},
		},
		{
			Description: "Simple query with basic filter(name), no results",
			Request: `query {
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
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic filter(age)",
		Request: `query {
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
				}`),
				(`{
				"Name": "Bob",
				"Age": 32
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
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple query with basic filter(age), greater than",
			Request: `query {
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
					"Age": 19
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
			Request: `query {
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
			Request: `query {
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
	test := testUtils.RequestTestCase{
		Description: "Simple query with logical compound filter (and)",
		Request: `query {
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
	test := testUtils.RequestTestCase{
		Description: "Simple query with logical compound filter (or)",
		Request: `query {
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
	test := testUtils.RequestTestCase{
		Description: "Simple query with special filter (or)",
		Request: `query {
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
