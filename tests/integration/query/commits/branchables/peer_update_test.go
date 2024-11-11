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
				NodeID: immutable.Some(1),
				Doc: `{
					"name":	"Chris"
				}`,
			},
			testUtils.WaitForSync{},
			// Note: node 1 does not recieve the first update from node 0 as it occured before the nodes were connected
			// node 0 has it as it recieved it when recieving the second update from node 1.  The cids and blocks remain
			// consistent across both nodes (minus the missing commits).
			testUtils.Request{
				NodeID: immutable.Some(0),
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
							"cid": testUtils.NewUniqueCid("collection, update2"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("collection, node1 update1"),
								},
								{
									"cid": testUtils.NewUniqueCid("doc, update2"),
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
							"cid": testUtils.NewUniqueCid("name, update2"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("name, node1 update1"),
								},
							},
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
							"cid": testUtils.NewUniqueCid("doc, update2"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("doc, node1 update1"),
								},
								{
									"cid": testUtils.NewUniqueCid("name, update2"),
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
				NodeID: immutable.Some(1),
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
							"cid": testUtils.NewUniqueCid("collection, update2"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("collection, node1 update1"),
								},
								{
									"cid": testUtils.NewUniqueCid("doc, update2"),
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
							"cid": testUtils.NewUniqueCid("name, update2"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("name, node1 update1"),
								},
							},
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
							"cid":   testUtils.NewUniqueCid("name, create"),
							"links": []map[string]any{},
						},
						{
							"cid": testUtils.NewUniqueCid("doc, update2"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("doc, node1 update1"),
								},
								{
									"cid": testUtils.NewUniqueCid("name, update2"),
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
							"name": "Chris",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
