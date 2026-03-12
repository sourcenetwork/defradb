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

package one_to_many

import (
	"testing"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestView_OneToManyWithTransformOnOuter(t *testing.T) {
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
			&action.AddLens{
				Lens: model.Lens{
					// This transform will copy the value from `name` into the `fullName` field,
					// like an overly-complicated alias
					Lenses: []model.LensModule{
						{
							Path: lenses.CopyModulePath,
							Arguments: map[string]any{
								"src": "name",
								"dst": "fullName",
							},
						},
					},
				},
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
						fullName: String
						books: [BookView]
					}
					interface BookView {
						name: String
					}
				`,
				TransformCID: immutable.Some("{{.LensID0}}"),
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Ferdowsi"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "Shahnameh",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: `
					query {
						AuthorView {
							fullName
							books {
								name
							}
						}
					}
				`,
				Results: map[string]any{
					"AuthorView": []map[string]any{
						{
							"fullName": "Ferdowsi",
							"books": []map[string]any{
								{
									"name": "Shahnameh",
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_OneToManyWithTransformAddingInnerDocs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Author {
						name: String
					}
				`,
			},
			&action.AddLens{
				Lens: model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.SetDefaultModulePath,
							Arguments: map[string]any{
								"dst": "books",
								"value": []map[string]any{
									{
										"name": "The Tragedy of Sohrab and Rostam",
									},
									{
										"name": "The Legend of Seyavash",
									},
								},
							},
						},
					},
				},
			},
			&action.AddView{
				Query: `
					Author {
						name
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
				TransformCID: immutable.Some("{{.LensID0}}"),
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Ferdowsi"
				}`,
			},
			&action.Request{
				Request: `
					query {
						AuthorView {
							name
							books {
								name
							}
						}
					}
				`,
				Results: map[string]any{
					"AuthorView": []map[string]any{
						{
							"name": "Ferdowsi",
							"books": []map[string]any{
								{
									"name": "The Tragedy of Sohrab and Rostam",
								},
								{
									"name": "The Legend of Seyavash",
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
