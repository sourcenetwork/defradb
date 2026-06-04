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
			&action.AddCollection{
				SDL: `
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
			&action.AddView{
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
			&action.PatchCollection{
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
			&action.AddDoc{
				// Set the `name` field only
				Doc: `{
					"firstName": "John",
					"lastName":  "S"
				}`,
			},
			&action.Request{
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
