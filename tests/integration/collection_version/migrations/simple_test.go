// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package migrations

import (
	"testing"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

// Migrations need to be able to be registered for unknown schema ids, so they
// may migrate to/from them if recieved by the P2P system.
func TestSchemaMigrationDoesNotErrorGivenUnknownSchemaRoots(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "does not exist",
					DestinationCollectionVersionID: "also does not exist",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "verified",
									"value": false,
								},
							},
						},
					},
				},
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{
					{
						VersionID:      "also does not exist",
						IsMaterialized: true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: "does not exist",
							Transform:          immutable.Some("{{.LensID0}}"),
						}),
					},
					{
						VersionID:      "does not exist",
						IsMaterialized: true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaMigrationGetMigrationsReturnsMultiple(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "does not exist",
					DestinationCollectionVersionID: "also does not exist",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "verified",
									"value": false,
								},
							},
						},
					},
				},
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "bafyreigsld6ten2pppcu2tgkbexqwdndckp6zt2vfjhuuheykqkgpmwk7i",
					DestinationCollectionVersionID: "bafyreigqfjat435ghyt66tdaucp7oi2mke5jafx3jw3rozanopihr2vf44",
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
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{
					{
						VersionID:      "also does not exist",
						IsMaterialized: true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: "does not exist",
							Transform:          immutable.Some("{{.LensID0}}"),
						}),
					},
					{
						IsMaterialized: true,
						VersionID:      "bafyreigqfjat435ghyt66tdaucp7oi2mke5jafx3jw3rozanopihr2vf44",
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: "bafyreigsld6ten2pppcu2tgkbexqwdndckp6zt2vfjhuuheykqkgpmwk7i",
							Transform:          immutable.Some("{{.LensID1}}"),
						}),
					},
					{
						IsMaterialized: true,
						VersionID:      "bafyreigsld6ten2pppcu2tgkbexqwdndckp6zt2vfjhuuheykqkgpmwk7i",
					},
					{
						VersionID:      "does not exist",
						IsMaterialized: true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaMigrationReplacesExistingMigationBasedOnSourceID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "a",
					DestinationCollectionVersionID: "b",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "verified",
									"value": false,
								},
							},
						},
					},
				},
			},
			testUtils.ConfigureMigration{
				// Replace the original migration with a new configuration
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "a",
					DestinationCollectionVersionID: "c",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "age",
									"value": 123,
								},
							},
						},
					},
				},
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{
					{
						VersionID:      "a",
						IsMaterialized: true,
					},
					{
						VersionID:      "b",
						IsMaterialized: true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: "a",
							Transform:          immutable.Some("{{.LensID0}}"),
						}),
					},
					{
						VersionID:      "c",
						IsMaterialized: true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: "a",
							Transform:          immutable.Some("{{.LensID1}}"),
						}),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
func TestSchemaMigration_ConfigureMigrationSkippingVersion_Errors(t *testing.T) {
	version1 := "bafyreihuyovjl5ezgpud5xyqnouzsgx25x3ssrx3ncdv5p3guocc3laqna"
	version3 := "bafyreih3uwvq6u5yqt65os3u5jdrrmy6gfi7wjq3vwvnm45jhjodbablhe"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users { }
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "Boolean"} }
					]
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "String"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      version1,
					DestinationCollectionVersionID: version3,
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "verified",
									"value": false,
								},
							},
						},
					},
				},
				ExpectedError: "cannot migrate between non-adjacent collection versions",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
