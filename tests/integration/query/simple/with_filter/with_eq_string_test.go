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

func TestQuerySimpleWithStringFilterBlock(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic filter (Name)",
		Request: `query {
					Users(filter: {Name: {_eq: "John"}}) {
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
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"Name": "John",
					"Age":  int64(21),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithStringEqualsNilFilterBlock(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic string nil filter",
		Request: `query {
					Users(filter: {Name: {_eq: null}}) {
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
					"Age": 60
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"Name": nil,
					"Age":  int64(60),
				},
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
						Users(filter: {Name: {_eq: "John"}}) {
							Name
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
				},
			},
			Results: map[string]any{
				"Users": []map[string]any{
					{
						"Name": "John",
					},
				},
			},
		},
		{
			Description: "Simple query with basic filter and selection (diff from filter)",
			Request: `query {
						Users(filter: {Name: {_eq: "John"}}) {
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
				},
			},
			Results: map[string]any{
				"Users": []map[string]any{
					{
						"Age": int64(21),
					},
				},
			},
		},
		{
			Description: "Simple query with basic filter(name), no results",
			Request: `query {
						Users(filter: {Name: {_eq: "Bob"}}) {
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
				},
			},
			Results: map[string]any{
				"Users": []map[string]any{},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}
