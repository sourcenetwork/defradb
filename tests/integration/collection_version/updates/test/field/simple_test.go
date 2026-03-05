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

func TestCollectionVersionUpdatesTestFieldNameErrors(t *testing.T) {
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
						{ "op": "test", "path": "/Users/Fields/1/name", "value": "Email" }
					]
				`,
				ExpectedError: "failed: test failed",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesTestFieldNamePasses(t *testing.T) {
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
						{ "op": "test", "path": "/Users/Fields/1/Name", "value": "name" }
					]
				`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesTestFieldErrors(t *testing.T) {
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
						{ "op": "test", "path": "/Users/Fields/1", "value": {"Name": "name", "Kind": 11} }
					]
				`,
				ExpectedError: "test failed",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesTestFieldPasses(t *testing.T) {
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
						{ "op": "test", "path": "/Users/Fields/1", "value": {
							"FieldID": "bafyreiaezc5g33yzhyzcgbyiv476lovyztyoliotzksdfogoep5ktgpedq",
							"Name": "name", "Kind": 11, "Typ":1, "RelationName": null, "IsPrimary": false, "DefaultValue": null, "Size": 0
						} }
					]
				`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesTestFieldPasses_UsingFieldNameAsIndex(t *testing.T) {
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
						{ "op": "test", "path": "/Users/Fields/name", "value": {
							"FieldID": "bafyreiaezc5g33yzhyzcgbyiv476lovyztyoliotzksdfogoep5ktgpedq",
							"Name": "name", "Kind": 11, "Typ":1, "RelationName": null, "IsPrimary": false, "DefaultValue": null, "Size": 0
						} }
					]
				`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesTestFieldPasses_TargettingKindUsingFieldNameAsIndex(t *testing.T) {
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
						{ "op": "test", "path": "/Users/Fields/name/Kind", "value": 11 }
					]
				`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
