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

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestView_OneToOneWithTransformOnOuter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddView{
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
