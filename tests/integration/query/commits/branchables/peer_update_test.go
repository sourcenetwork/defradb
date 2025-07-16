// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package branchables

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsBranchables_HandlesConcurrentUpdatesAcrossPeerConnection(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	collectionNode0Update3Cid := testUtils.NewSameValue()
	collectionNode1Update2Cid := testUtils.NewSameValue()
	collectionNode1Update1Cid := testUtils.NewSameValue()
	docNode0Update3Cid := testUtils.NewSameValue()
	collectionCreateCid := testUtils.NewSameValue()
	docNode1Update1Cid := testUtils.NewSameValue()
	docCreateCid := testUtils.NewSameValue()
	collectionNode0Update1Cid := testUtils.NewSameValue()
	docNode1Update2Cid := testUtils.NewSameValue()
	docNode0Update1Cid := testUtils.NewSameValue()
	nameNode0Update3Cid := testUtils.NewSameValue()
	nameNode1Update1Cid := testUtils.NewSameValue()
	nameNode1Update2Cid := testUtils.NewSameValue()
	nameNode0Update1Cid := testUtils.NewSameValue()
	nameCreateCid := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users @branchable {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"Fred"
				}`,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(1),
				Doc: `{
					"name":	"Shahzad"
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        0,
				CollectionIDs: []int{0},
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.WaitForSync{},
			testUtils.UpdateDoc{
				// Update node 1 after the peer connection has been established, this will cause the `Shahzad` commit
				// to be synced to node 0, as well as the related collection commits.
				NodeID: immutable.Some(1),
				Doc: `{
					"name":	"Chris"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.UpdateDoc{
				// Update node 0 after `Chris` and `Shahzad` have synced to node 0.  As this update happens after the peer
				// connection has been established, this will cause the `Fred` and `Addo` doc commits, and their corresponding
				// collection-level commits to sync to node 1.
				//
				// Now, all nodes should have a full history, including the 'offline' changes made before establishing the
				// peer connection.
				NodeID: immutable.Some(0),
				Doc: `{
				  "name": "Addo"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				// Strong eventual consistency must now have been established across both nodes, the result of this query
				// *must* exactly match across both nodes.
				Request: `query {
						commits {
							cid
							links {
								cid
							}
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": gomega.And(collectionNode0Update3Cid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": collectionNode1Update2Cid,
								},
								{
									"cid": collectionNode1Update1Cid,
								},
								{
									"cid": docNode0Update3Cid,
								},
							},
						},
						{
							"cid": gomega.And(collectionNode1Update1Cid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": collectionCreateCid,
								},
								{
									"cid": docNode1Update1Cid,
								},
							},
						},
						{
							"cid": gomega.And(collectionCreateCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": docCreateCid,
								},
							},
						},
						{
							"cid": gomega.And(collectionNode1Update2Cid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": collectionNode0Update1Cid,
								},
								{
									"cid": docNode1Update2Cid,
								},
							},
						},
						{
							"cid": gomega.And(collectionNode0Update1Cid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": collectionCreateCid,
								},
								{
									"cid": docNode0Update1Cid,
								},
							},
						},
						{
							"cid": gomega.And(nameNode0Update3Cid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": nameNode1Update1Cid,
								},
								{
									"cid": nameNode1Update2Cid,
								},
							},
						},
						{
							"cid": gomega.And(nameNode1Update2Cid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": nameNode0Update1Cid,
								},
							},
						},
						{
							"cid": gomega.And(nameNode0Update1Cid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": nameCreateCid,
								},
							},
						},
						{
							"cid":   gomega.And(nameCreateCid, uniqueCid),
							"links": []map[string]any{},
						},
						{
							"cid": gomega.And(nameNode1Update1Cid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": nameCreateCid,
								},
							},
						},
						{
							"cid": gomega.And(docNode0Update3Cid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": docNode1Update2Cid,
								},
								{
									"cid": docNode1Update1Cid,
								},
								{
									"cid": nameNode0Update3Cid,
								},
							},
						},
						{
							"cid": gomega.And(docNode1Update1Cid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": docCreateCid,
								},
								{
									"cid": nameNode1Update1Cid,
								},
							},
						},
						{
							"cid": gomega.And(docCreateCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": nameCreateCid,
								},
							},
						},
						{
							"cid": gomega.And(docNode1Update2Cid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": docNode0Update1Cid,
								},
								{
									"cid": nameNode1Update2Cid,
								},
							},
						},
						{
							"cid": gomega.And(docNode0Update1Cid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": docCreateCid,
								},
								{
									"cid": nameNode0Update1Cid,
								},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Addo",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
