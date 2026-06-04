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

func TestQuerySimpleWithBoolNotEqualsTrueFilterBlock(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Fred",
					"Verified": false
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Verified: {_neq: true}}) {
						Name
					}
				}`,
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
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithBoolNotEqualsNilFilterBlock(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Fred",
					"Verified": false
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Verified: {_neq: null}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Fred",
						},
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

func TestQuerySimpleWithBoolNotEqualsFalseFilterBlock(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Fred",
					"Verified": false
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Verified: {_neq: false}}) {
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
