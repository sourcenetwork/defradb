// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestView_OneToOneSameSchema(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one view with same schema",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type LeftHand {
						name: String
						holding: RightHand @primary @relation(name: "left_right")
						heldBy: RightHand @relation(name: "right_left")
					}
					type RightHand {
						name: String
						holding: LeftHand @primary @relation(name: "right_left")
						heldBy: LeftHand @relation(name: "left_right")
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					LeftHand {
						name
						heldBy {
							name
						}
					}
				`,
				// todo - such a setup appears to work, yet prevents the querying of `RightHand`s as the primary return object
				// thought - although, perhaps if the view is defined as such, Left and right hands *could* be merged by us into a single table
				SDL: `
					type HandView {
						name: String
						holding: HandView @primary
						heldBy: HandView
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Left hand 1"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":       "Right hand 1",
					"holding_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.Request{
				Request: `
					query {
						HandView {
							name
							heldBy {
								name
							}
						}
					}
				`,
				Results: map[string]any{
					"HandView": []map[string]any{
						{
							"name": "Left hand 1",
							"heldBy": map[string]any{
								"name": "Right hand 1",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_OneToOneEmbeddedSchemaIsNotLostOnNextUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one view followed by GQL type update",
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
			testUtils.CreateView{
				Query: `
					Author {
						name
						books {
							name
						}
					}
				`,
				SDL: `
					type AuthorView {
						name: String
						books: [BookView]
					}
					interface BookView {
						name: String
					}
				`,
			},
			// After creating the view, update the system's types again and ensure
			// that `BookView` is not forgotten.  A GQL error would appear if this
			// was broken as `AuthorView.books` would reference a type that does
			// not exist.
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
