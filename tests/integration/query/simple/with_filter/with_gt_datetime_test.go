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

func TestQuerySimpleWithDateTimeGTFilterBlockWithEqualValue(t *testing.T) {
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
					"CreatedAt": "2010-07-23T03:46:56-05:00"
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {CreatedAt: {_gt: "2017-07-20T03:46:56-05:00"}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithDateTimeGTFilterBlockWithGreaterValue(t *testing.T) {
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
					"CreatedAt": "2010-07-23T03:46:56-05:00"
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {CreatedAt: {_gt: "2017-07-22T03:46:56-05:00"}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithDateTimeGTFilterBlockWithLesserValue(t *testing.T) {
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
					"CreatedAt": "2010-07-23T03:46:56-05:00"
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {CreatedAt: {_gt: "2017-07-25T03:46:56-05:00"}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithDateTimeGTFilterBlockWithNilValue(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"CreatedAt": "2010-07-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob"
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {CreatedAt: {_gt: null}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithNilDateTimeGTAndNonNilFilterBlock_ShouldSucceed(t *testing.T) {
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
					Users(filter: {CreatedAt: {_gt: "2016-07-23T03:46:56-05:00"}}) {
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
