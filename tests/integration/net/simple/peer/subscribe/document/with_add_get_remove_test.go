// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package subscribe_test

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestP2PDocumentAddRemoveGetSingle(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToDocument{
				NodeID: 1,
				DocIDs: []testUtils.ColDocIndex{
					testUtils.NewColDocIndex(0, 0),
				},
			},
			testUtils.UnsubscribeToDocument{
				NodeID: 1,
				DocIDs: []testUtils.ColDocIndex{
					testUtils.NewColDocIndex(0, 0),
				},
			},
			testUtils.GetAllP2PDocuments{
				NodeID:         1,
				ExpectedDocIDs: []testUtils.ColDocIndex{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PDocumentAddRemoveGetMultiple(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
				},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToDocument{
				NodeID: 1,
				DocIDs: []testUtils.ColDocIndex{
					testUtils.NewColDocIndex(0, 0),
					testUtils.NewColDocIndex(0, 1),
				},
			},
			testUtils.UnsubscribeToDocument{
				NodeID: 1,
				DocIDs: []testUtils.ColDocIndex{
					testUtils.NewColDocIndex(0, 0),
				},
			},
			testUtils.GetAllP2PDocuments{
				NodeID: 1,
				ExpectedDocIDs: []testUtils.ColDocIndex{
					testUtils.NewColDocIndex(0, 1),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
