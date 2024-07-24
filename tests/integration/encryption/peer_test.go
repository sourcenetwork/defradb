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

/*func TestDocEncryptionPeer_IfPeerHasNoKey_ShouldNotFetch(t *testing.T) {
	test := testUtils.TestCase{
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
				Request: `query {
					Users {
						age
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}*/

func TestDocEncryptionPeer_UponSync_ShouldSyncEncryptedDAG(t *testing.T) {
	test := testUtils.TestCase{
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
							collectionID
							delta
							docID
							fieldId
							fieldName
							height
							links {
								cid
								name
							}
						}
					}
				`,
				Results: []map[string]any{
					{
						"cid":          "bafyreibdjepzhhiez4o27srv33xcd52yr336tpzqtkv36rdf3h3oue2l5m",
						"collectionID": int64(1),
						"delta":        encrypt(testUtils.CBORValue(21), john21DocID, ""),
						"docID":        john21DocID,
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafyreihkiua7jpwkye3xlex6s5hh2azckcaljfi2h3iscgub5sikacyrbu",
						"collectionID": int64(1),
						"delta":        encrypt(testUtils.CBORValue("John"), john21DocID, ""),
						"docID":        john21DocID,
						"fieldId":      "2",
						"fieldName":    "name",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafyreidxdhzhwjrv5s4x6cho5drz6xq2tc7oymzupf4p4gfk6eelsnc7ke",
						"collectionID": int64(1),
						"delta":        nil,
						"docID":        john21DocID,
						"fieldId":      "C",
						"fieldName":    nil,
						"height":       int64(1),
						"links": []map[string]any{
							{
								"cid":  "bafyreibdjepzhhiez4o27srv33xcd52yr336tpzqtkv36rdf3h3oue2l5m",
								"name": "age",
							},
							{
								"cid":  "bafyreihkiua7jpwkye3xlex6s5hh2azckcaljfi2h3iscgub5sikacyrbu",
								"name": "name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
