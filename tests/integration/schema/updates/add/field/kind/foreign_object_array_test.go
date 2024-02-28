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

func TestSchemaUpdatesAddFieldKindForeignObjectArray(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17)",
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 17} }
					]
				`,
				ExpectedError: "a `Schema` [name] must be provided when adding a new relation field. Field: foo, Kind: 17",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObjectArray_InvalidSchemaJson(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17), invalid schema json",
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 17, "Schema": 123} }
					]
				`,
				ExpectedError: "json: cannot unmarshal number into Go struct field SchemaFieldDescription.Fields.Schema of type string",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObjectArray_MissingRelationName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17), missing relation name",
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
							"Name": "foo", "Kind": 17, "Schema": "Users"
						}}
					]
				`,
				ExpectedError: "missing relation name. Field: foo",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObjectArray_IDFieldMissingKind(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17), id field missing kind",
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
							"Name": "foo", "Kind": 16, "IsPrimaryRelation": true, "Schema": "Users", "RelationName": "foo"
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

func TestSchemaUpdatesAddFieldKindForeignObjectArray_IDFieldInvalidKind(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17), id field invalid kind",
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
							"Name": "foo", "Kind": 16, "IsPrimaryRelation": true, "Schema": "Users", "RelationName": "foo"
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

func TestSchemaUpdatesAddFieldKindForeignObjectArray_IDFieldMissingRelationName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17), id field missing relation name",
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
							"Name": "foo", "Kind": 16, "IsPrimaryRelation": true, "Schema": "Users", "RelationName": "foo"
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

func TestSchemaUpdatesAddFieldKindForeignObjectArray_OnlyHalfRelationDefined(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17), only half relation defined",
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
							"Name": "foo", "Kind": 16, "IsPrimaryRelation": true, "Schema": "Users", "RelationName": "foo"
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

func TestSchemaUpdatesAddFieldKindForeignObjectArray_NoPrimaryDefined(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17), no primary defined",
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
							"Name": "foo", "Kind": 16, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": 17, "Schema": "Users", "RelationName": "foo"
						}}
					]
				`,
				ExpectedError: "primary side of relation not defined. RelationName: foo",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObjectArray_PrimaryDefinedOnManySide(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17), no primary defined",
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
							"Name": "foo", "Kind": 16,  "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1,  "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": 17, "IsPrimaryRelation": true, "Schema": "Users", "RelationName": "foo"
						}}
					]
				`,
				ExpectedError: "cannot set the many side of a relation as primary. Field: foobar",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObjectArray_Succeeds(t *testing.T) {
	key1 := "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"

	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17), valid, functional",
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
							"Name": "foo", "Kind": 16, "IsPrimaryRelation": true, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": 17, "Schema": "Users", "RelationName": "foo"
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
						"foobar": []map[string]any{},
					},
					{
						"name": "John",
						"foo":  nil,
						"foobar": []map[string]any{
							{
								"name": "Keenan",
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObjectArray_SinglePrimaryObjectKindSubstitution(t *testing.T) {
	key1 := "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"

	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17), with single object Kind substitution",
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
							"Name": "foo", "Kind": "Users", "IsPrimaryRelation": true, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": 17, "Schema": "Users", "RelationName": "foo"
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
						"foobar": []map[string]any{},
					},
					{
						"name": "John",
						"foo":  nil,
						"foobar": []map[string]any{
							{
								"name": "Keenan",
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObjectArray_SingleSecondaryObjectKindSubstitution(t *testing.T) {
	key1 := "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"

	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17), with single object Kind substitution",
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
							"Name": "foo", "Kind": 16, "IsPrimaryRelation": true, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": "[Users]", "Schema": "Users", "RelationName": "foo"
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
						"foo_id": "%s"
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
						"foobar": []map[string]any{},
					},
					{
						"name": "John",
						"foo":  nil,
						"foobar": []map[string]any{
							{
								"name": "Keenan",
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObjectArray_ObjectKindSubstitution(t *testing.T) {
	key1 := "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"

	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17), with object Kind substitution",
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
							"Name": "foo", "Kind": "Users", "IsPrimaryRelation": true, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": "[Users]", "Schema": "Users", "RelationName": "foo"
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
						"foobar": []map[string]any{},
					},
					{
						"name": "John",
						"foo":  nil,
						"foobar": []map[string]any{
							{
								"name": "Keenan",
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObjectArray_ObjectKindSubstitutionWithAutoSchemaValues(t *testing.T) {
	key1 := "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"

	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17), with object Kind substitution",
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
							"Name": "foobar", "Kind": "[Users]", "RelationName": "foo"
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
						"foobar": []map[string]any{},
					},
					{
						"name": "John",
						"foo":  nil,
						"foobar": []map[string]any{
							{
								"name": "Keenan",
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObjectArray_PrimaryObjectKindAndSchemaMismatch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17), with Kind and Schema mismatch",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaUpdate{
				Schema: `
					type Dog {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo", "Kind": "Users", "IsPrimaryRelation": true, "Schema": "Dog", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": "[Users]", "Schema": "Users", "RelationName": "foo"
						}}
					]
				`,
				ExpectedError: "field Kind does not match field Schema. Kind: Users, Schema: Dog",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObjectArray_SecondaryObjectKindAndSchemaMismatch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17), with Kind and Schema mismatch",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaUpdate{
				Schema: `
					type Dog {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo", "Kind": "Users", "IsPrimaryRelation": true, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": "[Users]", "Schema": "Dog", "RelationName": "foo"
						}}
					]
				`,
				ExpectedError: "field Kind does not match field Schema. Kind: [Users], Schema: Dog",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObjectArray_MissingPrimaryIDField(t *testing.T) {
	key1 := "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"

	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17), with auto id field generation",
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
							"Name": "foobar", "Kind": "[Users]", "RelationName": "foo"
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
						foo_id
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
						"name":   "Keenan",
						"foo_id": key1,
						"foo": map[string]any{
							"name": "John",
						},
						"foobar": []map[string]any{},
					},
					{
						"name":   "John",
						"foo":    nil,
						"foo_id": nil,
						"foobar": []map[string]any{
							{
								"name": "Keenan",
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObjectArray_MissingPrimaryIDField_DoesNotCreateIdOnManySide(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17), with auto id field generation",
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
							"Name": "foobar", "Kind": "[Users]", "RelationName": "foo"
						}}
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						foobar_id
					}
				}`,
				ExpectedError: `Cannot query field "foobar_id" on type "Users"`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
