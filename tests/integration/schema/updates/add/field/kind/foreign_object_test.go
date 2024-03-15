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

func TestSchemaUpdatesAddFieldKindForeignObject_MissingRelationName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object, missing relation name",
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
						}}
					]
				`,
				ExpectedError: "missing relation name. Field: foo",
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
							"Name": "foo", "Kind": "Users","IsPrimaryRelation": true, "RelationName": "foo"
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
							"Name": "foo", "Kind": "Users", "IsPrimaryRelation": true, "RelationName": "foo"
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

func TestSchemaUpdatesAddFieldKindForeignObject_IDFieldMissingRelationName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object, id field missing relation name",
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
							"Name": "foo", "Kind": "Users", "IsPrimaryRelation": true, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo_id", "Kind": 1} }
					]
				`,
				ExpectedError: "missing relation name. Field: foo_id",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObject_OnlyHalfRelationDefined(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object, only half relation defined",
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
							"Name": "foo", "Kind": "Users", "IsPrimaryRelation": true, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationName": "foo"
						}}
					]
				`,
				ExpectedError: "relation must be defined on both schemas. Field: foo, Type: Users",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObject_NoPrimaryDefined(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object, no primary defined",
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
							"Name": "foo", "Kind": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": "Users", "RelationName": "foo"
						}}
					]
				`,
				ExpectedError: "primary side of relation not defined. RelationName: foo",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObject_BothSidesPrimary(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object, both sides primary",
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
							"Name": "foo", "Kind": "Users", "IsPrimaryRelation": true, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": "Users", "IsPrimaryRelation": true, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar_id", "Kind": 1, "RelationName": "foo"
						}}
					]
				`,
				ExpectedError: "both sides of a relation cannot be primary. RelationName: foo",
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
							"Name": "foo", "Kind": "Users", "IsPrimaryRelation": true, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar_id", "Kind": 1, "RelationName": "foo"
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
						foobar {
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
						"foobar": nil,
					},
					{
						"name": "John",
						"foo":  nil,
						"foobar": map[string]any{
							"name": "Keenan",
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObject_MissingPrimaryIDField(t *testing.T) {
	key1 := "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"

	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object (16), with auto primary ID field creation",
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
							"Name": "foo", "Kind": "Users", "IsPrimaryRelation": true, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar_id", "Kind": 1, "RelationName": "foo"
						}}
					]
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: fmt.Sprintf(`{
						"name": "Keenan",
						"foo": "%s"
					}`,
					key1,
				),
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						foo {
							name
						}
						foobar {
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
						"foobar": nil,
					},
					{
						"name": "John",
						"foo":  nil,
						"foobar": map[string]any{
							"name": "Keenan",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObject_MissingSecondaryIDField(t *testing.T) {
	key1 := "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"

	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object (16), with auto secondary ID field creation",
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
							"Name": "foo", "Kind": "Users", "IsPrimaryRelation": true, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": "Users", "RelationName": "foo"
						}}
					]
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: fmt.Sprintf(`{
						"name": "Keenan",
						"foo": "%s"
					}`,
					key1,
				),
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						foo {
							name
						}
						foobar {
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
						"foobar": nil,
					},
					{
						"name": "John",
						"foo":  nil,
						"foobar": map[string]any{
							"name": "Keenan",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
