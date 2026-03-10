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

package field

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersionUpdatesCopyFieldErrors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "copy", "from": "/Users/Fields/1", "path": "/Users/Fields/2" }
					]
				`,
				ExpectedError: "moving fields is not currently supported. Name: email",
			},
			&action.Request{
				Request: `query {
					Users {
						name
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

func TestCollectionVersionUpdatesCopyFieldErrorsMultiple(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "copy", "from": "/Users/Fields/1", "path": "/Users/Fields/2" },
						{ "op": "copy", "from": "/Users/Fields/1", "path": "/Users/Fields/2" }
					]
				`,
				ExpectedError: "moving fields is not currently supported. Name: email, ProposedIndex: 2, ExistingIndex: 1\nmoving fields is not currently supported. Name: email, ProposedIndex: 3, ExistingIndex: 1\nmoving fields is not currently supported. Name: name, ProposedIndex: 4, ExistingIndex: 2\nduplicate field. Name: email\nduplicate field. Name: email",
			},
			&action.Request{
				Request: `query {
					Users {
						name
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

func TestCollectionVersionUpdatesCopyFieldWithAndReplaceName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			&action.PatchCollection{
				// Here we esentially use Email as a template, copying it and renaming the
				// clone.
				Patch: `
					[
						{ "op": "copy", "from": "/Users/Fields/1", "path": "/Users/Fields/3" },
						{ "op": "replace", "path": "/Users/Fields/3/Name", "value": "fax" },
						{ "op": "remove", "path": "/Users/Fields/3/FieldID" }
					]
				`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						email
						fax
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

// This is an odd test, but still a possibility and we should still cover it.
func TestCollectionVersionUpdatesCopyFieldWithReplaceNameAndKindSubstitution(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				// Here we esentially use Name as a template, copying it, and renaming and
				// re-typing the clone.
				Patch: `
					[
						{ "op": "copy", "from": "/Users/Fields/1", "path": "/Users/Fields/2" },
						{ "op": "replace", "path": "/Users/Fields/2/Name", "value": "age" },
						{ "op": "replace", "path": "/Users/Fields/2/Kind", "value": "Int" },
						{ "op": "remove", "path": "/Users/Fields/2/FieldID" }
					]
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 3
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							// It is important to test this with data, to ensure the type has been substituted correctly
							"age": int64(3),
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// This is an odd test, but still a possibility and we should still cover it.
func TestCollectionVersionUpdatesCopyFieldAndReplaceNameAndInvalidKindSubstitution(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				// Here we esentially use Name as a template, copying it and renaming and
				// re-typing the clone.
				Patch: `
					[
						{ "op": "copy", "from": "/Users/Fields/1", "path": "/Users/Fields/2" },
						{ "op": "replace", "path": "/Users/Fields/2/Name", "value": "Age" },
						{ "op": "replace", "path": "/Users/Fields/2/Kind", "value": "NotAValidKind" }
					]
				`,
				ExpectedError: "no type found for given name. Field: Age, Kind: NotAValidKind",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
