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
)

func TestEncryptedIndexAddPeer_SchemaWithEncryptedIndex_ShouldGenerateGQL(t *testing.T) {
	test := testUtils.TestCase{
		KMS:                        testUtils.KMS{Activated: true},
		EnableSearchableEncryption: true,
		Actions: []any{
			// Add peers to enable p2p so that SE gql queries are generated
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int @encryptedIndex
					}
				`,
			},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						encrypted_User(filter: {age: {_eq: 33}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"encrypted_User": gomega.Not(gomega.BeEmpty()),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestEncryptedIndexAddPeer_AfterAddRequest_ShouldGenerateGQL(t *testing.T) {
	test := testUtils.TestCase{
		KMS:                        testUtils.KMS{Activated: true},
		EnableSearchableEncryption: true,
		Actions: []any{
			// Add peers to enable p2p so that SE gql queries are generated
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int
					}
				`,
			},
			testUtils.AddEncryptedIndex{
				FieldName: "age",
			},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						encrypted_User(filter: {age: {_eq: 33}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"encrypted_User": gomega.Not(gomega.BeEmpty()),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
