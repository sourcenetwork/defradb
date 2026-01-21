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
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestQuerySimpleWithBoolNotEqualsTrueFilterBlock(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob"
				}`,
			},
			testUtils.CreateDoc{
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
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithBoolNotEqualsNilFilterBlock(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob"
				}`,
			},
			testUtils.CreateDoc{
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
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob"
				}`,
			},
			testUtils.CreateDoc{
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
			},
		},
	}

	executeTestCase(t, test)
}
