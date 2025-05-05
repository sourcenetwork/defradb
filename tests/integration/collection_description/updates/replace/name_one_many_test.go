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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColDescrUpdateReplaceNameOneToMany(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John Grisham",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "Painted House",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafkreifusksmaa5pzrz3s7l5empvmixh6vgekpbztuqc2jwpmvsagwy75a/Name",
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
