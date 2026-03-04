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

package copy

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersionUpdatesCopyCollectionWithRemoveIDAndReplaceName(t *testing.T) {
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
				// Here we esentially use Users as a template, copying it and renaming the
				// clone. It is deliberately blocked for now, but should function at somepoint.
				Patch: `
					[
						{ "op": "copy", "from": "/Users", "path": "/Book" },
						{ "op": "remove", "path": "/Book/CollectionID" },
						{ "op": "remove", "path": "/Book/VersionID" },
						{ "op": "replace", "path": "/Book/Name", "value": "Book" }
					]
				`,
				ExpectedError: "adding collections via patch is not supported. Name: Book",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
