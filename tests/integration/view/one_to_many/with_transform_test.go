// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many

import (
	"testing"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestView_OneToManyWithTransformOnOuter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many view with transform on outer",
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
						fullName: String
						books: [BookView]
					}
					interface BookView {
						name: String
					}
				`,
				Transform: immutable.Some(model.Lens{
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
				}),
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Ferdowsi"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "Shahnameh",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.Request{
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
				Results: []map[string]any{
					{
						"fullName": "Ferdowsi",
						"books": []any{
							map[string]any{
								"name": "Shahnameh",
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
		Description: "One to many view with transform adding inner docs",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Author {
						name: String
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					Author {
						name
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
				Transform: immutable.Some(model.Lens{
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
				}),
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Ferdowsi"
				}`,
			},
			testUtils.Request{
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
				Results: []map[string]any{
					{
						"name": "Ferdowsi",
						"books": []any{
							map[string]any{
								"name": "The Tragedy of Sohrab and Rostam",
							},
							map[string]any{
								"name": "The Legend of Seyavash",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
