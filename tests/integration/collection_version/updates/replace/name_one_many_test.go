// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

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
