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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesAddFieldKindNillableStringArray(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind nillable string array (21)",
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
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Foo", "Kind": 21} }
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

func TestSchemaUpdatesAddFieldKindNillableStringArrayWithCreate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind nillable string array (21) with create",
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
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Foo", "Kind": 21} }
					]
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"Name": "John",
					"Foo": ["hello", "پدر سگ", null]
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
						"Foo": []immutable.Option[string]{
							immutable.Some("hello"),
							immutable.Some("پدر سگ"),
							immutable.None[string](),
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
