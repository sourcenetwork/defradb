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

// Reproduces https://github.com/sourcenetwork/defradb/issues/4787.
// A subscriber peer with both a P2P collection subscription and a GraphQL
// subscription must receive the decrypted document when an encrypted doc
// is added on the source peer.
func TestDocEncryptionPeer_WithSubscription_ShouldDeliverDecryptedDoc(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			addUserCollection(),
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			&action.SubscriptionRequest{
				NodeID: immutable.Some(1),
				Request: `subscription {
					Users {
						age
					}
				}`,
				Results: []map[string]any{
					{
						"Users": []map[string]any{
							{
								"age": int64(21),
							},
						},
					},
				},
			},
			&action.AddDoc{
				NodeID:         immutable.Some(0),
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
