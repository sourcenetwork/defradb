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
	"github.com/stretchr/testify/require"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestColDescrUpdateReplaceCollectionSourceTransform(t *testing.T) {
	transformCfgJson, err := json.Marshal(
		model.Lens{
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
	)
	require.NoError(t, err)

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
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
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.PatchCollection{
				Patch: fmt.Sprintf(`
						[
							{ "op": "replace", "path": "/2/Sources/0/Transform", "value": %s }
						]
					`,
					transformCfgJson,
				),
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				// Without the new transform, `"Shahzad"` would have been returned
				Results: []map[string]any{
					{
						"name": "Fred",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
