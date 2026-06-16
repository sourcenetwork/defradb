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

	"github.com/onsi/gomega"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocEncryptionPeer_UponSync_ShouldSyncEncryptedDAG(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()
	nameCid := testUtils.NewSameValue()
	ageCid := testUtils.NewSameValue()

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
			&action.AddDoc{
				NodeID:         immutable.Some(0),
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `
					query {
						_commits {
							cid
							delta
							docID
							fieldName
							height
							links {
								cid
								fieldName
							}
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":       gomega.And(ageCid, uniqueCid),
							"delta":     encryptedCBORValueWithKey(testUtils.CBORValue(21), localDocIDKey(1, 1), ""),
							"docID":     testUtils.NewDocIndex(0, 0),
							"fieldName": "age",
							"height":    int64(1),
							"links":     []map[string]any{},
						},
						{
							"cid":       gomega.And(nameCid, uniqueCid),
							"delta":     encryptedCBORValueWithKey(testUtils.CBORValue("John"), localDocIDKey(1, 1), ""),
							"docID":     testUtils.NewDocIndex(0, 0),
							"fieldName": "name",
							"height":    int64(1),
							"links":     []map[string]any{},
						},
						{
							"cid":       uniqueCid,
							"delta":     nil,
							"docID":     testUtils.NewDocIndex(0, 0),
							"fieldName": "_C",
							"height":    int64(1),
							"links": []map[string]any{
								{
									"cid":       nameCid,
									"fieldName": "name",
								},
								{
									"cid":       ageCid,
									"fieldName": "age",
								},
							},
						},
					},
				},
				NonOrderedResults: true,
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
			addUserCollection(),
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			&action.AddDoc{
				NodeID:         immutable.Some(0),
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
			testUtils.WaitForSync{},
			// Do not wait for the key sync and request the document as soon as the dag has synced
			// The document will be returned if the key-sync has taken place already, if not, the set will
			// be empty.
			&action.Request{
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
