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

// This test ensures that nearby relation fields are not failing validation during a collection patch.
func TestCollectionVersionUpdatesAddField_DoesNotAffectExistingRelation(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						author: Author
					}

					type Author {
						name: String
						books: [Book]
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Book/Fields/-", "value": {"Name": "rating", "Kind": 4} }
					]
				`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
