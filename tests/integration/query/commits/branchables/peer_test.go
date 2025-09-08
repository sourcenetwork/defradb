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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsBranchables_SyncsAcrossPeerConnection(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	collectionCid := testUtils.NewSameValue()
	compositeCid := testUtils.NewSameValue()
	ageCid := testUtils.NewSameValue()
	nameCid := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
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
							"cid": gomega.And(collectionCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": compositeCid,
								},
							},
						},
						{
							"cid":   gomega.And(ageCid, uniqueCid),
							"links": []map[string]any{},
						},
						{
							"cid":   gomega.And(nameCid, uniqueCid),
							"links": []map[string]any{},
						},
						{
							"cid": gomega.And(compositeCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": ageCid,
								},
								{
									"cid": nameCid,
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
	uniqueCid := testUtils.NewUniqueValue()

	collectionDoc2CreateCid := testUtils.NewSameValue()
	collectionDoc1CreateCid := testUtils.NewSameValue()
	doc2CreateCid := testUtils.NewSameValue()
	doc1CreateCid := testUtils.NewSameValue()
	doc1NameCid := testUtils.NewSameValue()
	doc1AgeCid := testUtils.NewSameValue()
	doc2NameCid := testUtils.NewSameValue()
	doc2AgeCid := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
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
							"cid": gomega.And(collectionDoc2CreateCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": collectionDoc1CreateCid,
								},
								{
									"cid": doc2CreateCid,
								},
							},
						},
						{
							"cid": gomega.And(collectionDoc1CreateCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": doc1CreateCid,
								},
							},
						},
						{
							"cid":   gomega.And(doc2NameCid, uniqueCid),
							"links": []map[string]any{},
						},
						{
							"cid":   gomega.And(doc2AgeCid, uniqueCid),
							"links": []map[string]any{},
						},
						{
							"cid": gomega.And(doc2CreateCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": doc2NameCid,
								},
								{
									"cid": doc2AgeCid,
								},
							},
						},
						{
							"cid":   gomega.And(doc1NameCid, uniqueCid),
							"links": []map[string]any{},
						},
						{
							"cid":   gomega.And(doc1AgeCid, uniqueCid),
							"links": []map[string]any{},
						},
						{
							"cid": gomega.And(doc1CreateCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": doc1NameCid,
								},
								{
									"cid": doc1AgeCid,
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
