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
		Description: "Test schema update, add field with kind foreign object (16)",
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
				ExpectedError: "a `Schema` [name] must be provided when adding a new relation field. Field: foo, Kind: 16",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObject_InvalidSchemaJson(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object (16), invalid schema json",
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 16, "Schema": 123} }
					]
				`,
				ExpectedError: "json: cannot unmarshal number into Go struct field FieldDescription.Fields.Schema of type string",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObject_MissingRelationType(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object (16), missing relation type",
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 16, "Schema": "Users"} }
					]
				`,
				ExpectedError: "invalid RelationType. Field: foo, Expected: 1 and 4 or 8, with optionally 128, Actual: 0",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObject_UnknownSchema(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object (16), unknown schema",
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
							"Name": "foo", "Kind": 16, "RelationType": 5, "Schema": "Unknown"
						}}
					]
				`,
				ExpectedError: "no schema found for given name. Field: foo, Schema: Unknown",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObject_MissingRelationName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object (16), missing relation name",
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
							"Name": "foo", "Kind": 16, "RelationType": 5, "Schema": "Users"
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
		Description: "Test schema update, add field with kind foreign object (16), id field missing kind",
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
							"Name": "foo", "Kind": 16, "RelationType": 133, "Schema": "Users", "RelationName": "foo"
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
		Description: "Test schema update, add field with kind foreign object (16), id field invalid kind",
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
							"Name": "foo", "Kind": 16, "RelationType": 133, "Schema": "Users", "RelationName": "foo"
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

func TestSchemaUpdatesAddFieldKindForeignObject_IDFieldMissingRelationType(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object (16), id field missing relation type",
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
							"Name": "foo", "Kind": 16, "RelationType": 133, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo_id", "Kind": 1} }
					]
				`,
				ExpectedError: "invalid RelationType. Field: foo_id, Expected: 64, Actual: 0",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObject_IDFieldInvalidRelationType(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object (16), id field invalid RelationType",
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
							"Name": "foo", "Kind": 16, "RelationType": 133, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo_id", "Kind": 1, "RelationType": 4} }
					]
				`,
				ExpectedError: "invalid RelationType. Field: foo_id, Expected: 64, Actual: 4",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObject_IDFieldMissingRelationName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object (16), id field missing relation name",
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
							"Name": "foo", "Kind": 16, "RelationType": 133, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo_id", "Kind": 1, "RelationType": 64} }
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
		Description: "Test schema update, add field with kind foreign object (16), only half relation defined",
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
							"Name": "foo", "Kind": 16, "RelationType": 133, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
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
		Description: "Test schema update, add field with kind foreign object (16), no primary defined",
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
							"Name": "foo", "Kind": 16, "RelationType": 5, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": 16, "RelationType": 5, "Schema": "Users", "RelationName": "foo"
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
		Description: "Test schema update, add field with kind foreign object (16), both sides primary",
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
							"Name": "foo", "Kind": 16, "RelationType": 133, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": 16, "RelationType": 133, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar_id", "Kind": 1, "RelationType": 64, "Schema": "Users", "RelationName": "foo"
						}}
					]
				`,
				ExpectedError: "both sides of a relation cannot be primary. RelationName: foo",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObject_RelatedKindMismatch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object (16), related kind mismatch",
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
							"Name": "foo", "Kind": 16, "RelationType": 133, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": 17, "RelationType": 5, "Schema": "Users", "RelationName": "foo"
						}}
					]
				`,
				ExpectedError: "invalid Kind of the related field. RelationName: foo, Expected: 16, Actual: 17",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObject_RelatedRelationTypeMismatch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object (16), related relation type mismatch",
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
							"Name": "foo", "Kind": 16, "RelationType": 133, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": 16, "RelationType": 9, "Schema": "Users", "RelationName": "foo"
						}}
					]
				`,
				ExpectedError: "invalid RelationType of the related field. RelationName: foo, Expected: 4, Actual: 9",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObject_Succeeds(t *testing.T) {
	key1 := "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"

	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object (16), valid, functional",
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
							"Name": "foo", "Kind": 16, "RelationType": 133, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": 16, "RelationType": 5, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
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

func TestSchemaUpdatesAddFieldKindForeignObject_SinglePrimaryObjectKindSubstitution(t *testing.T) {
	key1 := "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"

	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object (16), with single object Kind substitution",
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
							"Name": "foo", "Kind": "Users", "RelationType": 133, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": 16, "RelationType": 5, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
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

func TestSchemaUpdatesAddFieldKindForeignObject_SingleSecondaryObjectKindSubstitution(t *testing.T) {
	key1 := "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"

	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object (16), with single object Kind substitution",
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
							"Name": "foo", "Kind": 16, "RelationType": 133, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": "Users", "RelationType": 5, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
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

func TestSchemaUpdatesAddFieldKindForeignObject_ObjectKindSubstitution(t *testing.T) {
	key1 := "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"

	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object (16), with object Kind substitution",
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
							"Name": "foo", "Kind": "Users", "RelationType": 133, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": "Users", "RelationType": 5, "Schema": "Users", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
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

func TestSchemaUpdatesAddFieldKindForeignObject_ObjectKindSubstitutionWithAutoSchemaValues(t *testing.T) {
	key1 := "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"

	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object (16), with object Kind substitution",
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
							"Name": "foo", "Kind": "Users", "RelationType": 133, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": "Users", "RelationType": 5, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
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

func TestSchemaUpdatesAddFieldKindForeignObject_ObjectKindAndSchemaMismatch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object (16), with Kind and Schema mismatch",
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
							"Name": "foo", "Kind": "Users", "RelationType": 133, "Schema": "Dog", "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": "Users", "RelationType": 5, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
						}}
					]
				`,
				ExpectedError: "field Kind does not match field Schema. Kind: Users, Schema: Dog",
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
							"Name": "foo", "Kind": "Users", "RelationType": 133, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": "Users", "RelationType": 5, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
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
							"Name": "foo", "Kind": "Users", "RelationType": 133, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo_id", "Kind": 1, "RelationType": 64, "RelationName": "foo"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foobar", "Kind": "Users", "RelationType": 5, "RelationName": "foo"
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
