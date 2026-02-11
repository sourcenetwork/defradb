// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package searchable_encryption

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestSEReplicator_IfDocCreatedWhileReplicatorIsOffline_ShouldRetry(t *testing.T) {
	test := testUtils.TestCase{
		EnableSearchableEncryption: true,
		SupportedDatabaseTypes: immutable.Some(
			[]state.DatabaseType{
				// This test only supports file type databases since it requires the ability to
				// stop and start a node without losing data.
				testUtils.BadgerFileType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type Users {
						name: String @encryptedIndex
						age: Int
					}
				`,
			},
			testUtils.CreateReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.Close{
				NodeID: immutable.Some(1),
			},
			&action.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			&action.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Fred",
					"age": 22
				}`,
			},
			testUtils.Start{
				NodeID: immutable.Some(1),
			},
			testUtils.WaitForSESync{},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					encrypted_Users(filter: {name: {_eq: "John"}}) {
						docIDs
					}
				}`,
				Results: map[string]any{
					"encrypted_Users": []map[string]any{
						{
							"docIDs": gomega.ConsistOf(testUtils.DocIDAt(0, 0)),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
