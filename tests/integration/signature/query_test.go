// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package signature

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/immutable"
)

func TestDocSignature_With_Should(t *testing.T) {
	test := testUtils.TestCase{
		EnabledBlockSigning: true,
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
                    type Users {
                        name: String
                        age: Int
                    }
                `},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
				IsDocEncrypted: true,
			},
			testUtils.Request{
				Request: `
                    query {
                        Users {
                            _docID
                            name
                            age
                        }
                    }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": testUtils.NewDocIndex(0, 0),
							"name":   "John",
							"age":    int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocSignature_IfDocIsPublic_ShouldFetchKeyAndDecrypt(t *testing.T) {
	test := testUtils.TestCase{
		EnabledBlockSigning: true,
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type User {
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
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"age": int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
