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

func TestCollectionMigrationGetMigrationsWithTxn(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.ConfigureMigration{
				TransactionID: immutable.Some(0),
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
				TransactionID: immutable.Some(0),
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
