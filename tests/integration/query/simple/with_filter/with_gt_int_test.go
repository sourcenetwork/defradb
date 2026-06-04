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

func TestQuerySimpleWithIntGreaterThanFilterBlock_ReturnOneAsOneMatches(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
					}`,
			},
			&action.AddDoc{
				Doc: `{
						"Name": "Bob",
						"Age": 19
					}`,
			},
			&action.Request{
				Request: `query {
						Users(filter: {Age: {_gt: 20}}) {
							Name
							Age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"Age":  int64(21),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithIntGreaterThanFilterBlock_ReturnNoneAsNoMatch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
					}`,
			},
			&action.AddDoc{
				Doc: `{
						"Name": "Bob",
						"Age": 32
					}`,
			},
			&action.Request{
				Request: `query {
						Users(filter: {Age: {_gt: 40}}) {
							Name
							Age
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

func TestQuerySimpleWithIntGreaterThanFilterBlock_ReturnAllMultiMatches(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
					}`,
			},
			&action.AddDoc{
				Doc: `{
						"Name": "Bob",
						"Age": 32
					}`,
			},
			&action.Request{
				Request: `query {
						Users(filter: {Age: {_gt: 20}}) {
							Name
							Age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
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
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithIntGreaterThanFilterBlockWithNullFilterValue(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob"
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Age: {_gt: null}}) {
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
