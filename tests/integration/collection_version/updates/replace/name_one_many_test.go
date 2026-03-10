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

func TestColVersionUpdateReplaceNameOneToMany(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Author {
						name: String
						books: [Book]
					}

					type Book {
						name: String
						author: Author
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John Grisham",
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "Painted House",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/Author/Name",
							"value": "Writer"
						}
					]
				`,
				ExpectedError: "collection name cannot be mutated. NewName: Writer, OldName: Author",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
