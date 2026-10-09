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

package replicator

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

func TestP2POneToOneReplicator_BranchableCollection_WithPatch(t *testing.T) {
	test := testUtils.TestCase{
		// The collection block arrives only by retry, which runs on the node sending
		// it. Releases before the fix in
		// https://github.com/sourcenetwork/defradb/pull/5307 don't retry it.
		OldFirstNodeSupportedFromVersion: "v1.2.0",
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
		}),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User @branchable {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/User/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.Close{
				NodeID: immutable.Some(1),
			},
			&action.AddDoc{
				// Create John on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.Start{
				NodeID: immutable.Some(1),
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
						_commits(
							filter: {fieldName: {_eq: null}}
						) {
							collectionVersionId
							docID
							fieldName
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"collectionVersionId": "bafyreibqvtb7gtuijcpobuld2hhtqmeltnfou6hclkc6gixeggywymk6ou",
							"docID":               nil,
							"fieldName":           nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
