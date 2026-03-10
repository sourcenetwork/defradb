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
					SourceCollectionVersionID:      "bafyreigbatez5rnojqa4ccfqsbguh4ckquxr76elgqij7ckftbxpwqniv4",
					DestinationCollectionVersionID: "bafyreihiiez4vcgh4rys2zfs74macgwyybchutjslyw2oin747enuywn54",
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
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			&action.PatchCollection{
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
			&action.Request{
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
