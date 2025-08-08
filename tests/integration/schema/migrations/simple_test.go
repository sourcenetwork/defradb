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
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

// Migrations need to be able to be registered for unknown schema ids, so they
// may migrate to/from them if recieved by the P2P system.
func TestSchemaMigrationDoesNotErrorGivenUnknownSchemaRoots(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema migration, unknown schema ids",
		Actions: []any{
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "does not exist",
					DestinationSchemaVersionID: "also does not exist",
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
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					IncludeInactive: immutable.Some(true),
				},
				ExpectedResults: []client.CollectionVersion{
					{
						VersionID:      "also does not exist",
						IsMaterialized: true,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: "does not exist",
								Transform: immutable.Some(
									model.Lens{
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
								),
							},
						},
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
		Description: "Test schema migration, multiple migrations",
		Actions: []any{
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "does not exist",
					DestinationSchemaVersionID: "also does not exist",
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
					SourceSchemaVersionID:      "bafyreigsld6ten2pppcu2tgkbexqwdndckp6zt2vfjhuuheykqkgpmwk7i",
					DestinationSchemaVersionID: "bafyreig2nfxuzl3cob7txuvybcct6mmsylt57oirzsrehffkho6bdxlvwy",
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
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					IncludeInactive: immutable.Some(true),
				},
				ExpectedResults: []client.CollectionVersion{
					{
						VersionID:      "also does not exist",
						IsMaterialized: true,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: "does not exist",
								Transform: immutable.Some(
									model.Lens{
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
								),
							},
						},
					},
					{
						IsMaterialized: true,
						VersionID:      "bafyreig2nfxuzl3cob7txuvybcct6mmsylt57oirzsrehffkho6bdxlvwy",
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: "bafyreigsld6ten2pppcu2tgkbexqwdndckp6zt2vfjhuuheykqkgpmwk7i",
								Transform: immutable.Some(
									model.Lens{
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
								),
							},
						},
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
		Description: "Test schema migration, replace migration",
		Actions: []any{
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "a",
					DestinationSchemaVersionID: "b",
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
					SourceSchemaVersionID:      "a",
					DestinationSchemaVersionID: "c",
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
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					IncludeInactive: immutable.Some(true),
				},
				ExpectedResults: []client.CollectionVersion{
					{
						VersionID:      "a",
						IsMaterialized: true,
					},
					{
						VersionID:      "b",
						IsMaterialized: true,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: "a",
								Transform: immutable.Some(
									model.Lens{
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
								),
							},
						},
					},
					{
						VersionID:      "c",
						IsMaterialized: true,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: "a",
								Transform: immutable.Some(
									model.Lens{
										Lenses: []model.LensModule{
											{
												Path: lenses.SetDefaultModulePath,
												Arguments: map[string]any{
													"dst":   "age",
													"value": float64(123),
												},
											},
										},
									},
								),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
