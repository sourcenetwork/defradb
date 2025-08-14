// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encryption

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocEncryptionPeer_UponSync_ShouldSyncEncryptedDAG(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			updateUserCollectionSchema(),
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID:         immutable.Some(0),
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `
					query {
						commits {
							cid
							delta
							docID
							fieldName
							height
							links {
								cid
								name
							}
						}
					}
				`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":       "bafyreia37txi77ajmma3t3o4hlnkx7qdbzymioplbuscla576i52rr5hri",
							"delta":     encrypt(testUtils.CBORValue(21), john21DocID, ""),
							"docID":     john21DocID,
							"fieldName": "age",
							"height":    int64(1),
							"links":     []map[string]any{},
						},
						{
							"cid":       "bafyreiajo5esphlst5bi2qjudn7uutk5layfa3edxb55rett54qj7gznai",
							"delta":     encrypt(testUtils.CBORValue("John"), john21DocID, ""),
							"docID":     john21DocID,
							"fieldName": "name",
							"height":    int64(1),
							"links":     []map[string]any{},
						},
						{
							"cid":       "bafyreicmnl5hzhq4q533a47igftavebkqjhxl22t3hag6yods5j6iydji4",
							"delta":     nil,
							"docID":     john21DocID,
							"fieldName": "_C",
							"height":    int64(1),
							"links": []map[string]any{
								{
									"cid":  "bafyreia37txi77ajmma3t3o4hlnkx7qdbzymioplbuscla576i52rr5hri",
									"name": "age",
								},
								{
									"cid":  "bafyreiajo5esphlst5bi2qjudn7uutk5layfa3edxb55rett54qj7gznai",
									"name": "name",
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionPeer_IfPeerDidNotReceiveKey_ShouldNotFetch(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			updateUserCollectionSchema(),
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID:         immutable.Some(0),
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
			testUtils.WaitForSync{},
			// Do not wait for the key sync and request the document as soon as the dag has synced
			// The document will be returned if the key-sync has taken place already, if not, the set will
			// be empty.
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						age
					}
				}`,
				Results: map[string]any{
					"Users": testUtils.AnyOf(
						// The key-sync has not yet completed
						[]map[string]any{},
						// The key-sync has completed
						[]map[string]any{
							{
								"age": int64(21),
							},
						},
					),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
