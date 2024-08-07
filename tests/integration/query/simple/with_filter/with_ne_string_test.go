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

func TestQuerySimpleWithStringNotEqualsFilterBlock(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with ne string filter",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Name: {_ne: "John"}}) {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(32),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithStringNotEqualsNilFilterBlock(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with ne string nil filter",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Age": 36
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Name: {_ne: null}}) {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(21),
						},
						{
							"Age": int64(32),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
