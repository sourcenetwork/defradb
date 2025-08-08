// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fields

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesRemoveField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, remove field",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "remove", "path": "/Users/Fields/2" }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				ExpectedError: "Cannot query field \"name\" on type \"Users\".",
			},
			testUtils.Request{
				Request: `query {
					Users {
						email
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

func TestSchemaUpdatesRemoveAllFields(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, remove all fields",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "remove", "path": "/Users/Fields" }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						_docID
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

func TestSchemaUpdatesRemoveFieldNameErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, remove field name",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "remove", "path": "/Users/Fields/2/Name" }
					]
				`,
				ExpectedError: "mutating an existing field is not supported. ProposedName: ",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesRemoveFieldKindErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, remove field kind",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "remove", "path": "/Users/Fields/2/Kind" }
					]
				`,
				ExpectedError: "mutating an existing field is not supported. ProposedName: ",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesRemoveFieldTypErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, remove field Typ",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "remove", "path": "/Users/Fields/2/Typ" }
					]
				`,
				ExpectedError: "mutating an existing field is not supported. ProposedName: name",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
