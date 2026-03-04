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

package add

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersionUpdatesAddSimpleErrorsAddingCollection(t *testing.T) {
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
						{ "op": "add", "path": "/-", "value": {"Name": "books"} }
					]
				`,
				ExpectedError: "adding collections via patch is not supported. Name: books",
			},
			&action.Request{
				Request: `query {
					Users {
						name
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

func TestCollectionVersionUpdatesAddSimpleErrorsAddingCollectionProp(t *testing.T) {
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
						{ "op": "add", "path": "/Users/-", "value": {"Foo": "Bar"} }
					]
				`,
				ExpectedError: `json: unknown field "-"`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddSimpleErrorsAddingUnsupportedCollectionProp(t *testing.T) {
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
						{ "op": "add", "path": "/Users/Foo/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
				ExpectedError: "add operation does not apply: doc is missing path",
			},
			&action.Request{
				Request: `query {
					Users {
						name
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

func TestCollectionVersionUpdatesAddSimpleErrorsAddingUnsupportedCollectionPropAtRoot(t *testing.T) {
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
						{ "op": "add", "path": "/Users/Foo/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
				ExpectedError: "add operation does not apply: doc is missing path",
			},
			&action.Request{
				Request: `query {
					Users {
						name
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
