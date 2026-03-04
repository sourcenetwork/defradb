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

func TestCollectionVersionUpdatesAddFieldSimple_FieldIndexedByName(t *testing.T) {
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
						{ "op": "add", "path": "/Users/Fields/email", "value": {"Kind": 11} }
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

func TestCollectionVersionUpdatesAddFieldSimple_FieldIndexedByNameWithSameNameDefinedInValue(t *testing.T) {
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
						{ "op": "add", "path": "/Users/Fields/email", "value": {"Name": "email","Kind": 11} }
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

func TestCollectionVersionUpdatesAddFieldSimple_FieldIndexedByNameWithDifferentNameDefinedInValue(t *testing.T) {
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
						{ "op": "add", "path": "/Users/Fields/email", "value": {"Name": "different field name","Kind": 11} }
					]
				`,
				ExpectedError: "the index used does not match the given name",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
func TestCollectionVersionUpdatesAddFieldSimple_FieldIndexedByNameMultipleTimes(t *testing.T) {
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
						{ "op": "add", "path": "/Users/Fields/email", "value": {"Kind": 11} },
						{ "op": "test", "path": "/Users/Fields/email/Kind", "value": 11 }
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
