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

func TestCollectionMigrationQueryInversesAcrossMultipleVersions(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
						height: Int
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }
					]
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "bafyreigbatez5rnojqa4ccfqsbguh4ckquxr76elgqij7ckftbxpwqniv4",
					DestinationCollectionVersionID: "bafyreigl262d3mo7rswhwwpymsioyr6d6stzxqqaf5vpkq4ahxrl6owrmu",
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
					SourceCollectionVersionID:      "bafyreigl262d3mo7rswhwwpymsioyr6d6stzxqqaf5vpkq4ahxrl6owrmu",
					DestinationCollectionVersionID: "bafyreihiiez4vcgh4rys2zfs74macgwyybchutjslyw2oin747enuywn54",
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
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 33,
					"height": 185
				}`,
			},
			testUtils.SetActiveCollectionVersion{
				VersionID: "bafyreigbatez5rnojqa4ccfqsbguh4ckquxr76elgqij7ckftbxpwqniv4",
			},
			&action.Request{
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
