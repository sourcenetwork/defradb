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

package replace

import (
	"testing"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/immutable"
)

func TestColVersionUpdateReplaceIsActive_GetCollectionsWithTxn(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				TransactionID: immutable.Some(0),
				SDL: `
					type Users {}
				`,
			},
			&action.PatchCollection{
				TransactionID: immutable.Some(0),
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafyreihuyovjl5ezgpud5xyqnouzsgx25x3ssrx3ncdv5p3guocc3laqna/IsActive",
							"value": false
						}
					]
				`,
			},
			&action.GetCollections{
				TransactionID: immutable.Some(0),
				FilterOptions: options.GetCollections().SetCollectionName("Users"),
			},
			&action.GetCollections{
				TransactionID: immutable.Some(0),
				FilterOptions: options.GetCollections().SetCollectionID("bafyreihuyovjl5ezgpud5xyqnouzsgx25x3ssrx3ncdv5p3guocc3laqna"),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
