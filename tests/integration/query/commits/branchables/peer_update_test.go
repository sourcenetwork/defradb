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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsBranchables_HandlesConcurrentUpdatesAcrossPeerConnection(t *testing.T) {
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
							"cid": testUtils.NewUniqueCid("collection, node0 update3"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("collection, node1 update2"),
								},
								{
									"cid": testUtils.NewUniqueCid("collection, node1 update1"),
								},
								{
									"cid": testUtils.NewUniqueCid("doc, node0 update3"),
								},
							},
						},
						{
							"cid": testUtils.NewUniqueCid("collection, node1 update1"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("collection, create"),
								},
								{
									"cid": testUtils.NewUniqueCid("doc, node1 update1"),
								},
							},
						},
						{
							"cid": testUtils.NewUniqueCid("collection, create"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("doc, create"),
								},
							},
						},
						{
							"cid": testUtils.NewUniqueCid("collection, node1 update2"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("collection, node0 update1"),
								},
								{
									"cid": testUtils.NewUniqueCid("doc, node1 update2"),
								},
							},
						},
						{
							"cid": testUtils.NewUniqueCid("collection, node0 update1"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("collection, create"),
								},
								{
									"cid": testUtils.NewUniqueCid("doc, node0 update1"),
								},
							},
						},
						{
							"cid": testUtils.NewUniqueCid("name, node0 update3"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("name, node1 update1"),
								},
								{
									"cid": testUtils.NewUniqueCid("name, node1 update2"),
								},
							},
						},
						{
							"cid": testUtils.NewUniqueCid("name, node1 update2"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("name, node0 update1"),
								},
							},
						},
						{
							"cid": testUtils.NewUniqueCid("name, node0 update1"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("name, create"),
								},
							},
						},
						{
							"cid":   testUtils.NewUniqueCid("name, create"),
							"links": []map[string]any{},
						},
						{
							"cid": testUtils.NewUniqueCid("name, node1 update1"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("name, create"),
								},
							},
						},
						{
							"cid": testUtils.NewUniqueCid("doc, node0 update3"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("doc, node1 update2"),
								},
								{
									"cid": testUtils.NewUniqueCid("doc, node1 update1"),
								},
								{
									"cid": testUtils.NewUniqueCid("name, node0 update3"),
								},
							},
						},
						{
							"cid": testUtils.NewUniqueCid("doc, node1 update1"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("doc, create"),
								},
								{
									"cid": testUtils.NewUniqueCid("name, node1 update1"),
								},
							},
						},
						{
							"cid": testUtils.NewUniqueCid("doc, create"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("name, create"),
								},
							},
						},
						{
							"cid": testUtils.NewUniqueCid("doc, node1 update2"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("doc, node0 update1"),
								},
								{
									"cid": testUtils.NewUniqueCid("name, node1 update2"),
								},
							},
						},
						{
							"cid": testUtils.NewUniqueCid("doc, node0 update1"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("doc, create"),
								},
								{
									"cid": testUtils.NewUniqueCid("name, node0 update1"),
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
