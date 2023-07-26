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

	"github.com/lens-vm/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

// Migrations need to be able to be registered for unknown schema ids, so they
// may migrate to/from them if recieved by the P2P system.
func TestSchemaMigrationDoesNotErrorGivenUnknownSchemaIDs(t *testing.T) {
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
			testUtils.GetMigrations{
				ExpectedResults: []client.LensConfig{
					{
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
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
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
					SourceSchemaVersionID:      "bafkreihn4qameldz3j7rfundmd4ldhxnaircuulk6h2vcwnpcgxl4oqffq",
					DestinationSchemaVersionID: "bafkreia56p6i6o3l4jijayiqd5eiijsypjjokbldaxnmqgeav6fe576hcy",
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
			testUtils.GetMigrations{
				ExpectedResults: []client.LensConfig{
					{
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
					{
						SourceSchemaVersionID:      "bafkreihn4qameldz3j7rfundmd4ldhxnaircuulk6h2vcwnpcgxl4oqffq",
						DestinationSchemaVersionID: "bafkreia56p6i6o3l4jijayiqd5eiijsypjjokbldaxnmqgeav6fe576hcy",
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
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
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
			testUtils.GetMigrations{
				ExpectedResults: []client.LensConfig{
					{
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
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}
