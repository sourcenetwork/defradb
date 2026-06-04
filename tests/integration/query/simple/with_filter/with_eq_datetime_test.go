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

func TestQuerySimpleWithDateTimeEqualsFilterBlock(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21,
					"CreatedAt": "2017-07-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32,
					"CreatedAt": "2016-07-23T03:46:56-05:00"
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {CreatedAt: {_eq: "2017-07-23T03:46:56-05:00"}}) {
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
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithDateTimeEqualsNilFilterBlock(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21,
					"CreatedAt": "2017-07-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32,
					"CreatedAt": "2016-07-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Fred",
					"Age": 44
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {CreatedAt: {_eq: null}}) {
						Name
						Age
						CreatedAt
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":      "Fred",
							"Age":       int64(44),
							"CreatedAt": nil,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithNilDateTimeEqualsAndNonNilFilterBlock_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				DocMap: map[string]any{
					"Name":      "John",
					"Age":       int64(21),
					"CreatedAt": "2017-07-23T03:46:56-05:00",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"Name":      "Bob",
					"Age":       int64(32),
					"CreatedAt": "2016-07-23T03:46:56-05:00",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"Name": "Fred",
					"Age":  44,
				},
			},
			&action.Request{
				Request: `query {
					Users(filter: {CreatedAt: {_eq: "2016-07-23T03:46:56-05:00"}}) {
						Name
						Age
						CreatedAt
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":      "Bob",
							"Age":       int64(32),
							"CreatedAt": testUtils.MustParseTime("2016-07-23T03:46:56-05:00"),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
