// Copyright 2024 Democratized Data Foundation
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
	"time"

	"github.com/onsi/gomega"
	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

const userCollectionGQLSchema = (`
	type Users {
		name: String
		age: Int @encryptedIndex
		verified: Boolean
	}
`)

const (
	john21Doc = `{
		"name":	"John",
		"age":	21
	}`
	islam33Doc = `{
		"name":	"Islam",
		"age":	33
	}`
)

func updateUserCollectionSchema() testUtils.SchemaUpdate {
	return testUtils.SchemaUpdate{
		Schema: userCollectionGQLSchema,
	}
}

func TestDocEncryptionPeer_UponSync_ShouldSyncEncryptedDAG(t *testing.T) {
	test := testUtils.TestCase{
		KMS:                        testUtils.KMS{Activated: true},
		EnableSearchableEncryption: true,
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			updateUserCollectionSchema(),
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
				SEEnabled:    true,
			},
			testUtils.CreateDoc{
				NodeID:         immutable.Some(0),
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
			testUtils.Wait{
				Duration: time.Millisecond * 100,
			},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						Users_encrypted(filter: {age: {_eq: 21}}) {
							docIDs
						}
					}
				`,
				Results: map[string]any{
					"Users_encrypted": []map[string]any{
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
