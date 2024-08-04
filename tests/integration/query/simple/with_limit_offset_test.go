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

func TestQuerySimpleWithLimit0(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with limit 0",
		Request: `query {
					Users(limit: 0) {
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
					"Name": "Bob",
				},
				{
					"Name": "John",
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithLimit(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple query with basic limit",
			Request: `query {
						Users(limit: 1) {
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
						"Name": "Bob",
						"Age":  int64(32),
					},
				},
			},
		},
		{
			Description: "Simple query with basic limit, more rows",
			Request: `query {
						Users(limit: 2) {
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
			Results: map[string]any{
				"Users": []map[string]any{
					{
						"Name": "Carlo",
						"Age":  int64(55),
					},
					{
						"Name": "Bob",
						"Age":  int64(32),
					},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQuerySimpleWithLimitAndOffset(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple query with basic limit & offset",
			Request: `query {
						Users(limit: 1, offset: 1) {
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
		},
		{
			Description: "Simple query with basic limit & offset, more rows",
			Request: `query {
						Users(limit: 2, offset: 2) {
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
			Results: map[string]any{
				"Users": []map[string]any{
					{
						"Name": "John",
						"Age":  int64(21),
					},
					{
						"Name": "Alice",
						"Age":  int64(19),
					},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQuerySimpleWithOffset(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple query with offset only",
			Request: `query {
						Users(offset: 1) {
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
		},
		{
			Description: "Simple query with offset only, more rows",
			Request: `query {
						Users(offset: 2) {
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
						"Name": "Melynda",
						"Age": 30
					}`,
				},
			},
			Results: map[string]any{
				"Users": []map[string]any{
					{
						"Name": "Melynda",
						"Age":  int64(30),
					},
					{
						"Name": "John",
						"Age":  int64(21),
					},
					{
						"Name": "Alice",
						"Age":  int64(19),
					},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}
