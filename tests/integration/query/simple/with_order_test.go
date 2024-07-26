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

func TestQuerySimpleWithEmptyOrder(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with empty order",
		Request: `query {
					Users(order: {}) {
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
				{
					"Name": "John",
					"Age":  int64(21),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithNumericOrderAscending(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic order ASC",
		Request: `query {
					Users(order: {Age: ASC}) {
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
					"Name": "Alice",
					"Age":  int64(19),
				},
				{
					"Name": "John",
					"Age":  int64(21),
				},
				{
					"Name": "Bob",
					"Age":  int64(32),
				},
				{
					"Name": "Carlo",
					"Age":  int64(55),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithDateTimeOrderAscending(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic order ASC",
		Request: `query {
					Users(order: {CreatedAt: ASC}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21,
					"CreatedAt": "2021-07-23T03:46:56-05:00"
				}`,
				`{
					"Name": "Bob",
					"Age": 32,
					"CreatedAt": "2032-07-23T03:46:56-05:00"
				}`,
				`{
					"Name": "Carlo",
					"Age": 55,
					"CreatedAt": "2055-07-23T03:46:56-05:00"
				}`,
				`{
					"Name": "Alice",
					"Age": 19,
					"CreatedAt": "2019-07-23T03:46:56-05:00"
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"Name": "Alice",
					"Age":  int64(19),
				},
				{
					"Name": "John",
					"Age":  int64(21),
				},
				{
					"Name": "Bob",
					"Age":  int64(32),
				},
				{
					"Name": "Carlo",
					"Age":  int64(55),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithNumericOrderDescending(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic order DESC",
		Request: `query {
					Users(order: {Age: DESC}) {
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
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithDateTimeOrderDescending(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic order DESC",
		Request: `query {
					Users(order: {CreatedAt: DESC}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21,
					"CreatedAt": "2021-07-23T03:46:56-05:00"
				}`,
				`{
					"Name": "Bob",
					"Age": 32,
					"CreatedAt": "2032-07-23T03:46:56-05:00"
				}`,
				`{
					"Name": "Carlo",
					"Age": 55,
					"CreatedAt": "2055-07-23T03:46:56-05:00"
				}`,
				`{
					"Name": "Alice",
					"Age": 19,
					"CreatedAt": "2019-07-23T03:46:56-05:00"
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
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithNumericOrderDescendingAndBooleanOrderAscending(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with compound order",
		Request: `query {
					Users(order: {Age: DESC, Verified: ASC}) {
						Name
						Age
						Verified
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21,
					"Verified": true
				}`,
				`{
					"Name": "Bob",
					"Age": 21,
					"Verified": false
				}`,
				`{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
				`{
					"Name": "Alice",
					"Age": 19,
					"Verified": false
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"Name":     "Carlo",
					"Age":      int64(55),
					"Verified": true,
				},
				{
					"Name":     "Bob",
					"Age":      int64(21),
					"Verified": false,
				},
				{
					"Name":     "John",
					"Age":      int64(21),
					"Verified": true,
				},
				{
					"Name":     "Alice",
					"Age":      int64(19),
					"Verified": false,
				},
			},
		},
	}

	executeTestCase(t, test)
}
