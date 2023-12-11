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

func TestSchemaUpdatesAddFieldKindBigInt(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind big int (8)",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 8} }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						foo
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindBigIntWithCreate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind big int (8) with create",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 8} }
					]
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"foo": "123456"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						foo
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
						"foo":  "123456",
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindBigIntSubstitutionWithCreate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind big int substitution with create",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": "Blob"} }
					]
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"foo": "123456"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						foo
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
						"foo":  "123456",
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
