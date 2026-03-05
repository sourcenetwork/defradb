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

package one_to_one

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestView_OneToOneSameSchema(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddView{
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
					type HandView @materialized(if: false) {
						name: String
						holding: HandView @primary
						heldBy: HandView
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Left hand 1"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":       "Right hand 1",
					"_holdingID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
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
			&action.AddView{
				Query: `
					Author {
						name
						books {
							name
						}
					}
				`,
				SDL: `
					type AuthorView @materialized(if: false) {
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
