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

// TODO: This test documents an unimplemented feature. Tracked by:
// https://github.com/sourcenetwork/defradb/issues/3212
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
							"cid": testUtils.NewUniqueCid("collection"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("composite"),
								},
							},
						},
						{
							"cid":   testUtils.NewUniqueCid("age"),
							"links": []map[string]any{},
						},
						{
							"cid":   testUtils.NewUniqueCid("name"),
							"links": []map[string]any{},
						},
						{
							"cid": testUtils.NewUniqueCid("composite"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("age"),
								},
								{
									"cid": testUtils.NewUniqueCid("name"),
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
						// Note: The collection commit has not synced.
						{
							"cid":   testUtils.NewUniqueCid("age"),
							"links": []map[string]any{},
						},
						{
							"cid":   testUtils.NewUniqueCid("name"),
							"links": []map[string]any{},
						},
						{
							"cid": testUtils.NewUniqueCid("composite"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("age"),
								},
								{
									"cid": testUtils.NewUniqueCid("name"),
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
