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

func TestCollectionVersionUpdatesMoveFieldErrors(t *testing.T) {
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
						{ "op": "move", "from": "/Users/Fields/1", "path": "/Users/Fields/-" }
					]
				`,
				ExpectedError: "moving fields is not currently supported. Name: name, ProposedIndex: 1, ExistingIndex: 2",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesMoveFieldErrorsMultiple(t *testing.T) {
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
						{ "op": "move", "from": "/Users/Fields/1", "path": "/Users/Fields/-" }
					]
				`,
				ExpectedError: "moving fields is not currently supported. Name: name, ProposedIndex: 1, ExistingIndex: 2\nmoving fields is not currently supported. Name: email, ProposedIndex: 2, ExistingIndex: 1",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
