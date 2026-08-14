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

package collection_version

import (
	"testing"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// These tests check that present-but-empty lens migration and absent
// migrations are both working distinctly, and properly.

const usersCollectionVersion1ID = "bafyreihuyovjl5ezgpud5xyqnouzsgx25x3ssrx3ncdv5p3guocc3laqna"

// This tests no migration
func TestCollectionVersionPatch_NoneMigration_NoTransformRecorded(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "String"} }
					]
				`,
				Lens: immutable.None[model.Lens](),
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetCollectionName("Users"),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsActive:       true,
						IsMaterialized: true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: usersCollectionVersion1ID,
						}),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This tests an empty migration
func TestCollectionVersionPatch_SomeEmptyMigration_TransformRecorded(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "String"} }
					]
				`,
				Lens: immutable.Some(model.Lens{}),
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetCollectionName("Users"),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsActive:       true,
						IsMaterialized: true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: usersCollectionVersion1ID,
							Transform:          immutable.Some("lensID1"),
						}),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
