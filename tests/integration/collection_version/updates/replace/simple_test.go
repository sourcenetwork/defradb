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

package replace

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersionUpdatesReplaceCollectionErrors(t *testing.T) {
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
				// Replace Users with Book
				Patch: `
					[
						{
							"op": "replace", "path": "/Users", "value": {
								"Name": "Book",
								"Fields": [
									{"Name": "name", "Kind": 11}
								]
							}
						}
					]
				`,
				// WARNING: An error is still expected if/when we allow the adding of collections, as this also
				// implies that the "Users" collection is to be deleted.  Only once we support the adding *and*
				// removal of collections should this not error.
				ExpectedError: "adding collections via patch is not supported. Name: Book",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
