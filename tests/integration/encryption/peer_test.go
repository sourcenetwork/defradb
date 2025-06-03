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
							"cid":       "bafyreiba7bxnqquldhojcnkak7afamaxssvjk4uav4ev4lwqgixarvvp4i",
							"delta":     encrypt(testUtils.CBORValue(21), john21DocID, ""),
							"docID":     john21DocID,
							"fieldName": "age",
							"height":    int64(1),
							"links":     []map[string]any{},
						},
						{
							"cid":       "bafyreigawlzc5zi2juad5vldnwvels5qsehymb45maoeamdbckajwcao24",
							"delta":     encrypt(testUtils.CBORValue("John"), john21DocID, ""),
							"docID":     john21DocID,
							"fieldName": "name",
							"height":    int64(1),
							"links":     []map[string]any{},
						},
						{
							"cid":       "bafyreidl77w6pex7uworttm5bsqyvli5qxqoqy3q2n2xqor5vrqfr3woee",
							"delta":     nil,
							"docID":     john21DocID,
							"fieldName": "_C",
							"height":    int64(1),
							"links": []map[string]any{
								{
									"cid":  "bafyreiba7bxnqquldhojcnkak7afamaxssvjk4uav4ev4lwqgixarvvp4i",
									"name": "age",
								},
								{
									"cid":  "bafyreigawlzc5zi2juad5vldnwvels5qsehymb45maoeamdbckajwcao24",
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
