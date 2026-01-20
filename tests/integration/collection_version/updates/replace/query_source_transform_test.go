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

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestColVersionUpdateReplaceQuerySourceTransform(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						firstName: String
						lastName: String
					}
				`,
			},
			&action.AddLens{
				Lens: model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.CopyModulePath,
							Arguments: map[string]any{
								"src": "firstName",
								"dst": "fullName",
							},
						},
					},
				},
			},
			&action.AddLens{
				Lens: model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.CopyModulePath,
							Arguments: map[string]any{
								"src": "lastName",
								"dst": "fullName",
							},
						},
					},
				},
			},
			&action.CreateView{
				Query: `
					User {
						firstName
						lastName
					}
				`,
				SDL: `
					type UserView @materialized(if: false) {
						fullName: String
					}
				`,
				TransformCID: immutable.Some("{{.LensID0}}"),
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/UserView/Query/Transform",
							"value": "{{.LensID1}}"
						}
					]
				`,
			},
			testUtils.CreateDoc{
				// Set the `name` field only
				Doc: `{
					"firstName": "John",
					"lastName":  "S"
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						UserView {
							fullName
						}
					}
				`,
				Results: map[string]any{
					"UserView": []map[string]any{
						{
							"fullName": "S",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
