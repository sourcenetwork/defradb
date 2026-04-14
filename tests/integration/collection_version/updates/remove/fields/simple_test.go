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

package fields

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestCollectionVersionUpdatesRemoveField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Removing a field from the schema version makes it unqueryable while other fields remain.",
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
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
						{ "op": "remove", "path": "/Users/Fields/2" }
					]
				`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				ExpectedError: "Cannot query field \"name\" on type \"Users\".",
			},
			&action.Request{
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

func TestCollectionVersionUpdatesRemoveAllFields(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Removing all fields from the schema version leaves only the built-in _docID queryable.",
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
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
						{ "op": "remove", "path": "/Users/Fields" }
					]
				`,
			},
			&action.Request{
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

func TestCollectionVersionUpdatesRemoveFieldNameErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Removing the Name property of an existing field returns a mutation error.",
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
						{ "op": "remove", "path": "/Users/Fields/2/Name" }
					]
				`,
				ExpectedError: "mutating an existing field is not supported. ProposedName: ",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesRemoveFieldKindErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Removing the Kind property of an existing field returns a mutation error.",
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
						{ "op": "remove", "path": "/Users/Fields/2/Kind" }
					]
				`,
				ExpectedError: "mutating an existing field is not supported. ProposedName: ",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesRemoveFieldTypErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Removing the Typ property of an existing field returns a mutation error.",
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
						{ "op": "remove", "path": "/Users/Fields/2/Typ" }
					]
				`,
				ExpectedError: "mutating an existing field is not supported. ProposedName: name",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
