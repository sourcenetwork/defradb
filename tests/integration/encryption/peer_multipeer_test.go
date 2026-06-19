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

package encryption

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// Node 0 holds the encrypted doc and key. Nodes 1, 2 are KMS-active bystanders
// on the encryption topic but are not subscribed to the collection, so they
// reply to fetch-key requests with an empty/error payload. Node 3 (requester)
// is connected to all three; without the fix it consumed only the first reply
// and the bystanders' empty replies routinely won the race.
// See https://github.com/sourcenetwork/defradb/issues/4779.
func TestDocEncryptionPeer_MultiplePeersOnEncryptionTopic_ShouldFetchKeyFromHolder(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			addUserCollection(),

			// Connect bystanders before the holder so they sit ahead of Node 0 in
			// Node 3's pubsub mesh, reliably winning the race to respond.
			testUtils.ConnectPeers{
				SourceNodeID: 3,
				TargetNodeID: 1,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 3,
				TargetNodeID: 2,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 3,
				TargetNodeID: 0,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        3,
				CollectionIDs: []int{0},
			},

			&action.AddDoc{
				NodeID:         immutable.Some(0),
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
			testUtils.WaitForSync{},

			&action.Request{
				NodeID: immutable.Some(3),
				Request: `query {
					Users {
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"age": int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
