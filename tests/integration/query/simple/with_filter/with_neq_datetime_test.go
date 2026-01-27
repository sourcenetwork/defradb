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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithDateTimeNotEqualsFilterBlock(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21,
					"CreatedAt": "2017-07-23T03:46:56-05:00"
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32,
					"CreatedAt": "2011-07-23T03:46:56-05:00"
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {CreatedAt: {_neq: "2017-07-23T03:46:56-05:00"}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithDateTimeNotEqualsNilFilterBlock(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21,
					"CreatedAt": "2017-07-23T03:46:56-05:00"
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32,
					"CreatedAt": "2011-07-23T03:46:56-05:00"
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Fred",
					"Age": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {CreatedAt: {_neq: null}}) {
						Name
					}
				}`,
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
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithNilDateTimeNotEqualAndNonNilFilterBlock_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				DocMap: map[string]any{
					"Name":      "John",
					"Age":       int64(21),
					"CreatedAt": "2017-07-23T03:46:56-05:00",
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"Name":      "Bob",
					"Age":       int64(32),
					"CreatedAt": "2016-07-23T03:46:56-05:00",
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"Name": "Fred",
					"Age":  44,
				},
			},
			&action.Request{
				Request: `query {
					Users(filter: {CreatedAt: {_neq: "2016-07-23T03:46:56-05:00"}}) {
						Name
						Age
						CreatedAt
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":      "John",
							"Age":       int64(21),
							"CreatedAt": testUtils.MustParseTime("2017-07-23T03:46:56-05:00"),
						},
						{
							"Name":      "Fred",
							"Age":       int64(44),
							"CreatedAt": nil,
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}
