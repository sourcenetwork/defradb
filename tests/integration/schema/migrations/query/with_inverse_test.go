// Copyright 2024 Democratized Data Foundation
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

	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestSchemaMigrationQueryInversesAcrossMultipleVersions(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema migration, inverses across multiple migrated versions",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						age: Int
						height: Int
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }
					]
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "bafyreih7useaapqn4pf6k5rxb2oufmsjb3e7xnccmbjr2njva3bgpdwyzu",
					DestinationSchemaVersionID: "bafyreih5sk2gsbddpvkil76x5sgbdgx6gmglhycpyvgli3szatqnfmxaka",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "age",
									"value": 30,
								},
							},
						},
					},
				},
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "bafyreih5sk2gsbddpvkil76x5sgbdgx6gmglhycpyvgli3szatqnfmxaka",
					DestinationSchemaVersionID: "bafyreicrtehvnvxkjdxac523mb7kkiwyn3wntj3yiuljskcco5ixhcpk7y",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "height",
									"value": 190,
								},
							},
						},
					},
				},
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 33,
					"height": 185
				}`,
			},
			testUtils.SetActiveCollectionVersion{
				SchemaVersionID: "bafyreih7useaapqn4pf6k5rxb2oufmsjb3e7xnccmbjr2njva3bgpdwyzu",
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						age
						height
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "John",
							"age":    nil,
							"height": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
