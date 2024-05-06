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
	"fmt"
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesAddFieldKindForeignObject(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object",
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 16} }
					]
				`,
				ExpectedError: "no type found for given name. Type: 16",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObject_UnknownSchema(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object, unknown schema",
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo", "Kind": "Unknown"
						}}
					]
				`,
				ExpectedError: "no type found for given name. Field: foo, Kind: Unknown",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObject_IDFieldMissingKind(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object, id field missing kind",
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo", "Kind": "Users"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo_id"} }
					]
				`,
				ExpectedError: "relational id field of invalid kind. Field: foo_id, Expected: ID, Actual: 0",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObject_IDFieldInvalidKind(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object, id field invalid kind",
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo", "Kind": "Users"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo_id", "Kind": 2} }
					]
				`,
				ExpectedError: "relational id field of invalid kind. Field: foo_id, Expected: ID, Actual: Boolean",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObject_Succeeds(t *testing.T) {
	key1 := "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"

	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object, valid, functional",
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo", "Kind": "Users"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1
						}}
					]
				`,
			},
			testUtils.Request{
				Request: `mutation {
						create_Users(input: {name: "John"}) {
							_docID
						}
					}`,
				Results: []map[string]any{
					{
						"_docID": key1,
					},
				},
			},
			testUtils.Request{
				Request: fmt.Sprintf(`mutation {
						create_Users(input: {name: "Keenan", foo: "%s"}) {
							name
							foo {
								name
							}
						}
					}`,
					key1,
				),
				Results: []map[string]any{
					{
						"name": "Keenan",
						"foo": map[string]any{
							"name": "John",
						},
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						foo {
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Keenan",
						"foo": map[string]any{
							"name": "John",
						},
					},
					{
						"name": "John",
						"foo":  nil,
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
