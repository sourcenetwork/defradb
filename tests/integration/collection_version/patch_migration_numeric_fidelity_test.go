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
	"github.com/sourcenetwork/defradb/tests/lenses"
)

// These tests illustrate that every client computes the same CID for the same large integers
// in tests with lens migration. Previously, values abbove 2^53 were being incorrectly rounded
// by somee clients during decoding. Therefore these tests exist as regression tests of this
// issuee.

const usersCollectionVersion1ID = "bafyreihuyovjl5ezgpud5xyqnouzsgx25x3ssrx3ncdv5p3guocc3laqna"

const largeIntArgumentLensCID = "bafyreihzes3vm5dqvvd6h6jxtdkgc4ts3eey4ixvkc3424kxnu66xt3ohe"

func TestCollectionVersionPatch_LargeIntegerLensArgument_ProducesConsistentCID(t *testing.T) {
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }
					]
				`,
				Lens: immutable.Some(model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.SetDefaultModulePath,
							Arguments: map[string]any{
								"dst":   "verified",
								"value": int64(9007199254740993), // 2^53 + 1, not representable exactly as float64
							},
						},
					},
				}),
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
							Transform:          immutable.Some(largeIntArgumentLensCID),
						}),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}