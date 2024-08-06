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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestColDescrUpdateReplaceQuerySourceTransform(t *testing.T) {
	newTransformCfgJson, err := json.Marshal(
		model.Lens{
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
	)
	require.NoError(t, err)

	test := testUtils.TestCase{
		Description: "Simple view with transform",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						firstName: String
						lastName: String
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					User {
						firstName
						lastName
					}
				`,
				SDL: `
					type UserView {
						fullName: String
					}
				`,
				Transform: immutable.Some(model.Lens{
					// This transform will copy the value from `firstName` into the `fullName` field,
					// like an overly-complicated alias
					Lenses: []model.LensModule{
						{
							Path: lenses.CopyModulePath,
							Arguments: map[string]any{
								"src": "firstName",
								"dst": "fullName",
							},
						},
					},
				}),
			},
			testUtils.PatchCollection{
				Patch: fmt.Sprintf(`
						[
							{ "op": "replace", "path": "/2/Sources/0/Transform", "value": %s }
						]
					`,
					newTransformCfgJson,
				),
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
