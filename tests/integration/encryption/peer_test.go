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
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":          "bafyreidkuvcdxxkyoeapnmttu6l2vk43qnm3zuzpxegbifpj6w24jrvrxq",
							"collectionID": int64(1),
							"delta":        encrypt(testUtils.CBORValue(21), john21DocID, ""),
							"docID":        john21DocID,
							"fieldId":      "1",
							"fieldName":    "age",
							"height":       int64(1),
							"links":        []map[string]any{},
						},
						{
							"cid":          "bafyreihdlv4fvvptetghxzyerxt4jc4zgprecybhoijrfjuyxqe55qw3x4",
							"collectionID": int64(1),
							"delta":        encrypt(testUtils.CBORValue("John"), john21DocID, ""),
							"docID":        john21DocID,
							"fieldId":      "2",
							"fieldName":    "name",
							"height":       int64(1),
							"links":        []map[string]any{},
						},
						{
							"cid":          "bafyreie5jegw4c2hg56bbiv6cgxmfz336jruukjakbjuyapockfnn6b5le",
							"collectionID": int64(1),
							"delta":        nil,
							"docID":        john21DocID,
							"fieldId":      "C",
							"fieldName":    nil,
							"height":       int64(1),
							"links": []map[string]any{
								{
									"cid":  "bafyreidkuvcdxxkyoeapnmttu6l2vk43qnm3zuzpxegbifpj6w24jrvrxq",
									"name": "age",
								},
								{
									"cid":  "bafyreihdlv4fvvptetghxzyerxt4jc4zgprecybhoijrfjuyxqe55qw3x4",
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
					"Users": testUtils.AnyOf{
						// The key-sync has not yet completed
						[]map[string]any{},
						// The key-sync has completed
						[]map[string]any{
							{
								"age": int64(21),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
