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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesAddFieldKindStringArray(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind string array (12)",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Foo", "Kind": 12} }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						Name
						Foo
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestSchemaUpdatesAddFieldKindStringArrayWithCreate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind string array (12) with create",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Foo", "Kind": 12} }
					]
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"Name": "John",
					"Foo": ["bar", "pub", "inn", "out", "hokey", "cokey", "pepsi", "beer", "bar", "pub", "..."]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						Name
						Foo
					}
				}`,
				Results: []map[string]any{
					{
						"Name": "John",
						"Foo":  []string{"bar", "pub", "inn", "out", "hokey", "cokey", "pepsi", "beer", "bar", "pub", "..."},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
