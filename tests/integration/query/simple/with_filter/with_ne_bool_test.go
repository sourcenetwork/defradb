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

func TestQuerySimpleWithBoolNotEqualsTrueFilterBlock(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with ne true filter",
		Request: `query {
					Users(filter: {Verified: {_ne: true}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Verified": true
				}`,
				`{
					"Name": "Bob"
				}`,
				`{
					"Name": "Fred",
					"Verified": false
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"Name": "Fred",
				},
				{
					"Name": "Bob",
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithBoolNotEqualsNilFilterBlock(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with ne nil filter",
		Request: `query {
					Users(filter: {Verified: {_ne: null}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Verified": true
				}`,
				`{
					"Name": "Bob"
				}`,
				`{
					"Name": "Fred",
					"Verified": false
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"Name": "John",
				},
				{
					"Name": "Fred",
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithBoolNotEqualsFalseFilterBlock(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with ne false filter",
		Request: `query {
					Users(filter: {Verified: {_ne: false}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Verified": true
				}`,
				`{
					"Name": "Bob"
				}`,
				`{
					"Name": "Fred",
					"Verified": false
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"Name": "John",
				},
				{
					"Name": "Bob",
				},
			},
		},
	}

	executeTestCase(t, test)
}
