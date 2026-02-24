// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package kind

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesAddFieldKindStringArray(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 12} }
					]
				`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						foo
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindStringArrayWithAdd(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 12} }
					]
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"foo": ["bar", "pub", "inn", "out", "hokey", "cokey", "pepsi", "beer", "bar", "pub", "..."]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						foo
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"foo":  []string{"bar", "pub", "inn", "out", "hokey", "cokey", "pepsi", "beer", "bar", "pub", "..."},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindStringArraySubstitutionWithAdd(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": "[String!]"} }
					]
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"foo": ["bar", "pub", "inn", "out", "hokey", "cokey", "pepsi", "beer", "bar", "pub", "..."]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						foo
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"foo":  []string{"bar", "pub", "inn", "out", "hokey", "cokey", "pepsi", "beer", "bar", "pub", "..."},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
