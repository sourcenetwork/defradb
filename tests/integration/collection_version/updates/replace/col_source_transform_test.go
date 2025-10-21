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

	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestColVersionUpdateReplaceCollectionSourceTransform(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// This ConfigureMigration action is a temporary work around as we have not yet exposed a
			// means to add lenses into Defra.  The collection ids are made up and have no impact on
			// the test.  The ID is passed into the next PatchCollection action.
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "bafyreih7useaapqn4pf6k5rxb2oufmsjb3e7xnccmbjr2njva3bgpdwyzu",
					DestinationSchemaVersionID: "bafyreiehn7n6uox2x4rjkiunezy2q3deom4keocn7riqpa5xa64c7gqx7u",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "name",
									"value": "Fred",
								},
							},
						},
					},
				},
			},
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/Users/PreviousVersion/Transform",
							"value": "{{.LensID0}}"
						}
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				// Without the new transform, `"Shahzad"` would have been returned
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
