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

package branchables

import (
	"slices"
	"testing"

	"github.com/onsi/gomega"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

type commitGraphNode struct {
	Links []string
	Heads []string
}

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
			&action.AddCollection{
				SDL: `
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
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
						_commits {
							cid
							links {
								cid
							}
							heads {
								cid
							}
						}
					}`,
				NonOrderedResults: true,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": gomega.And(collectionCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": compositeCid,
								},
							},
							"heads": []map[string]any{},
						},
						{
							"cid":   gomega.And(ageCid, uniqueCid),
							"links": []map[string]any{},
							"heads": []map[string]any{},
						},
						{
							"cid":   gomega.And(nameCid, uniqueCid),
							"links": []map[string]any{},
							"heads": []map[string]any{},
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
							"heads": []map[string]any{},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsBranchables_SyncsMultipleAcrossPeerConnection(t *testing.T) {
	var expectedGraph map[string]commitGraphNode

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
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
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"Fred",
					"age":	25
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
						_commits {
							cid
							links {
								cid
							}
							heads {
								cid
							}
						}
					}`,
				Asserter: action.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					graph := normalizeCommitGraph(t, result)
					assertTwoDocBranchableCreateGraph(t, graph)
					if expectedGraph == nil {
						expectedGraph = graph
						return true, ""
					}
					require.Equal(t, expectedGraph, graph)
					return true, ""
				}),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func normalizeCommitGraph(t testing.TB, result map[string]any) map[string]commitGraphNode {
	commits := action.ConvertToArrayOfMaps(t, result["_commits"])
	graph := make(map[string]commitGraphNode, len(commits))
	for _, commit := range commits {
		cid, ok := commit["cid"].(string)
		require.True(t, ok)
		require.NotContains(t, graph, cid)

		graph[cid] = commitGraphNode{
			Links: commitGraphCIDs(t, commit["links"]),
			Heads: commitGraphCIDs(t, commit["heads"]),
		}
	}
	return graph
}

func commitGraphCIDs(t testing.TB, value any) []string {
	links := action.ConvertToArrayOfMaps(t, value)
	cids := make([]string, len(links))
	for i, link := range links {
		cid, ok := link["cid"].(string)
		require.True(t, ok)
		cids[i] = cid
	}
	slices.Sort(cids)
	return cids
}

func assertTwoDocBranchableCreateGraph(t testing.TB, graph map[string]commitGraphNode) {
	require.Len(t, graph, 8)

	leafCIDs := make(map[string]struct{})
	for cid, node := range graph {
		if len(node.Links) == 0 && len(node.Heads) == 0 {
			leafCIDs[cid] = struct{}{}
		}
	}
	require.Len(t, leafCIDs, 4)

	compositeCIDs := make(map[string]struct{})
	for cid, node := range graph {
		if len(node.Links) != 2 || len(node.Heads) != 0 {
			continue
		}
		_, hasFirstLeaf := leafCIDs[node.Links[0]]
		_, hasSecondLeaf := leafCIDs[node.Links[1]]
		if hasFirstLeaf && hasSecondLeaf {
			compositeCIDs[cid] = struct{}{}
		}
	}
	require.Len(t, compositeCIDs, 2)

	var collectionRootCID string
	collectionHeadCount := 0
	collectionLinkedComposites := make(map[string]struct{})
	for cid, node := range graph {
		if len(node.Links) != 1 {
			continue
		}
		if _, ok := compositeCIDs[node.Links[0]]; !ok {
			continue
		}
		collectionLinkedComposites[node.Links[0]] = struct{}{}

		switch len(node.Heads) {
		case 0:
			require.Empty(t, collectionRootCID)
			collectionRootCID = cid
		case 1:
			collectionHeadCount++
		default:
			require.Failf(t, "unexpected collection heads", "cid %s has heads %v", cid, node.Heads)
		}
	}

	require.Len(t, collectionLinkedComposites, 2)
	require.NotEmpty(t, collectionRootCID)
	require.Equal(t, 1, collectionHeadCount)
}
