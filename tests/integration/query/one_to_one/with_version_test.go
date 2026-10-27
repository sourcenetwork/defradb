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

func TestQueryOneToOne_WithVersionOnOuterBeforeJoin(t *testing.T) {
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
						published: Book @primary
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "فارسی دوم دبستان"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "نمی دانم",
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: `
					query {
						Book {
							name
							author {
								name
							}
							_version {
								docID
							}
						}
					}
				`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "فارسی دوم دبستان",
							"_version": []map[string]any{
								{
									"docID": "{{.DocID0_0}}",
								},
							},
							"author": map[string]any{
								"name": "نمی دانم",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOne_WithVersionOnOuterAfterJoin(t *testing.T) {
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
						published: Book @primary
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "فارسی دوم دبستان"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "نمی دانم",
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: `
					query {
						Book {
							name
							_version {
								docID
							}
							author {
								name
							}
						}
					}
				`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "فارسی دوم دبستان",
							"_version": []map[string]any{
								{
									"docID": "{{.DocID0_0}}",
								},
							},
							"author": map[string]any{
								"name": "نمی دانم",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
