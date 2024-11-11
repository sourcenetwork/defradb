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

func TestQueryCommitsBranchables_SyncsAcrossPeerConnection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users @branchable {
						name: String
						age: Int
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
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
							"cid": testUtils.NewUniqueCid(0),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid(1),
								},
							},
						},
						{
							"cid":   testUtils.NewUniqueCid(2),
							"links": []map[string]any{},
						},
						{
							"cid":   testUtils.NewUniqueCid(3),
							"links": []map[string]any{},
						},
						{
							"cid": testUtils.NewUniqueCid(1),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid(2),
								},
								{
									"cid": testUtils.NewUniqueCid(3),
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

func TestQueryCommitsBranchables_SyncsMultipleAcrossPeerConnection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users @branchable {
						name: String
						age: Int
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"Fred",
					"age":	25
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
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
							"cid": testUtils.NewUniqueCid("collection, doc2 create"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("collection, doc1 create"),
								},
								{
									"cid": testUtils.NewUniqueCid("doc2 create"),
								},
							},
						},
						{
							"cid": testUtils.NewUniqueCid("collection, doc1 create"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("doc1 create"),
								},
							},
						},
						{
							"cid":   testUtils.NewUniqueCid("doc1 name"),
							"links": []map[string]any{},
						},
						{
							"cid":   testUtils.NewUniqueCid("doc1 age"),
							"links": []map[string]any{},
						},
						{
							"cid": testUtils.NewUniqueCid("doc1 create"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("doc1 name"),
								},
								{
									"cid": testUtils.NewUniqueCid("doc1 age"),
								},
							},
						},
						{
							"cid":   testUtils.NewUniqueCid("doc2 name"),
							"links": []map[string]any{},
						},
						{
							"cid":   testUtils.NewUniqueCid("doc2 age"),
							"links": []map[string]any{},
						},
						{
							"cid": testUtils.NewUniqueCid("doc2 create"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("doc2 name"),
								},
								{
									"cid": testUtils.NewUniqueCid("doc2 age"),
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
