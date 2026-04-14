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

package kind

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersionUpdatesAddFieldKindStringArray(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a non-nullable String array field succeeds and the field is queryable.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersionUpdatesAddFieldKindStringArrayWithAdd(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a non-nullable String array field and inserting a document stores and retrieves string values.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersionUpdatesAddFieldKindStringArraySubstitutionWithAdd(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a [String!] field using string kind substitution stores and retrieves the string array values.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
