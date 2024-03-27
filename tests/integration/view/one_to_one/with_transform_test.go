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

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestView_OneToOneWithTransformOnOuter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one view with transform on outer",
		Actions: []any{
			testUtils.SchemaUpdate{
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
			testUtils.CreateView{
				Query: `
					Author {
						name
						book {
							name
						}
					}
				`,
				SDL: `
					type AuthorView {
						fullName: String
						book: BookView
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
				Doc: `{
					"name":	"Shahnameh",
					"author": "bae-db3c6923-c6a4-5386-8301-b20a5454bf1d"
				}`,
			},
			testUtils.Request{
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
				Results: []map[string]any{
					{
						"fullName": "Ferdowsi",
						"book": map[string]any{
							"name": "Shahnameh",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
