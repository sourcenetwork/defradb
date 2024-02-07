// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package query

import (
	"testing"

	"github.com/lens-vm/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestSchemaMigrationQueryWithRestart(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema migration, with restart",
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
					"name": "John"
				}`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "bafkreibjb4h5nudsei7cq2kkontjinmjpbqls2tmowqp5nxougu4tuus4i",
					DestinationSchemaVersionID: "bafkreih6o2jyurelxtpbg66gk23pio2tq6o3aed334z6w2u3qwve3at7ku",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "verified",
									"value": true,
								},
							},
						},
					},
				},
			},
			testUtils.Restart{},
			testUtils.Request{
				Request: `query {
					Users {
						name
						verified
					}
				}`,
				Results: []map[string]any{
					{
						"name":     "John",
						"verified": true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
