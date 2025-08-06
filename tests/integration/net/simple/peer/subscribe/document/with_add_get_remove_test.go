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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestP2PDocumentAddRemoveGetSingle(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
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
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			testUtils.UnsubscribeToDocument{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			testUtils.GetAllP2PDocuments{
				NodeID:         1,
				ExpectedDocIDs: []state.ColDocIndex{},
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
			&action.AddSchema{
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
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
					state.NewColDocIndex(0, 1),
				},
			},
			testUtils.UnsubscribeToDocument{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			testUtils.GetAllP2PDocuments{
				NodeID: 1,
				ExpectedDocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 1),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
