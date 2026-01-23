// Copyright 2024 Democratized Data Foundation
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

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestView_OneToOneWithTransformOnOuter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Author {
						name: String
						book: Book
					}
					type Book {
						name: String
						author: Author @primary
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
			&action.CreateView{
				Query: `
					Author {
						name
						book {
							name
						}
					}
				`,
				SDL: `
					type AuthorView @materialized(if: false) {
						fullName: String
						book: BookView
					}
					interface BookView {
						name: String
					}
				`,
				TransformCID: immutable.Some("{{.LensID0}}"),
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
			&action.Request{
				Request: `
					query {
						AuthorView {
							fullName
							book {
								name
							}
						}
					}
				`,
				Results: map[string]any{
					"AuthorView": []map[string]any{
						{
							"fullName": "Ferdowsi",
							"book": map[string]any{
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
