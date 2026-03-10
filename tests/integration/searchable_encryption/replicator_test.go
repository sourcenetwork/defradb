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

package searchable_encryption

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestSEReplicator_IfDocAddedWhileReplicatorIsOffline_ShouldRetry(t *testing.T) {
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
			&action.AddCollection{
				SDL: `
					type Users {
						name: String @encryptedIndex
						age: Int
					}
				`,
			},
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.Close{
				NodeID: immutable.Some(1),
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			&action.AddDoc{
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
