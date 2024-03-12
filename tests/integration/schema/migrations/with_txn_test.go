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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestSchemaMigrationGetMigrationsWithTxn(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema migration, with txn",
		Actions: []any{
			testUtils.ConfigureMigration{
				TransactionID: immutable.Some(0),
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
				TransactionID: immutable.Some(0),
				FilterOptions: client.CollectionFetchOptions{
					IncludeInactive: immutable.Some(true),
				},
				ExpectedResults: []client.CollectionDescription{
					{
						ID:              1,
						SchemaVersionID: "does not exist",
					},
					{
						ID:              2,
						SchemaVersionID: "also does not exist",
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: 1,
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
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
