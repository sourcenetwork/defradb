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

package test

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersionUpdatesTestAddField(t *testing.T) {
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
				Patch: `
					[
						{ "op": "test", "path": "/Users/Name", "value": "Users" },
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
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
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddNonNillableBoolField_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String }`,
			},
			&action.PatchCollection{
				Patch: `
					[{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": 15} }]
				`,
				ExpectedError: "adding a non-nillable field to an existing collection is not supported",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddNonNillableIntField_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String }`,
			},
			&action.PatchCollection{
				Patch: `
					[{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "score", "Kind": 23} }]
				`,
				ExpectedError: "adding a non-nillable field to an existing collection is not supported",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddNonNillableFloat64Field_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String }`,
			},
			&action.PatchCollection{
				Patch: `
					[{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "rating", "Kind": 24} }]
				`,
				ExpectedError: "adding a non-nillable field to an existing collection is not supported",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddNonNillableFloat32Field_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String }`,
			},
			&action.PatchCollection{
				Patch: `
					[{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "weight", "Kind": 25} }]
				`,
				ExpectedError: "adding a non-nillable field to an existing collection is not supported",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddNonNillableStringField_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String }`,
			},
			&action.PatchCollection{
				Patch: `
					[{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 26} }]
				`,
				ExpectedError: "adding a non-nillable field to an existing collection is not supported",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddNonNillableDateTimeField_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String }`,
			},
			&action.PatchCollection{
				Patch: `
					[{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "createdAt", "Kind": 27} }]
				`,
				ExpectedError: "adding a non-nillable field to an existing collection is not supported",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddNonNillableBlobField_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String }`,
			},
			&action.PatchCollection{
				Patch: `
					[{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "avatar", "Kind": 28} }]
				`,
				ExpectedError: "adding a non-nillable field to an existing collection is not supported",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddNonNillableJSONField_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String }`,
			},
			&action.PatchCollection{
				Patch: `
					[{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "metadata", "Kind": 29} }]
				`,
				ExpectedError: "adding a non-nillable field to an existing collection is not supported",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesTestAddFieldBlockedByTest(t *testing.T) {
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
				Patch: `
					[
						{ "op": "test", "path": "/Users/Name", "value": "Author" },
						{ "op": "add", "path": "/Users/Fields/-", "value": {"name": "Email", "Kind": 11} }
					]
				`,
				ExpectedError: "test failed",
			},
			&action.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				ExpectedError: "Cannot query field \"email\" on type \"Users\"",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
