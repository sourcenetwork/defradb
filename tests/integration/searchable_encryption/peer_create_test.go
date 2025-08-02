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
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/immutable"
)

func TestEncryptedIndexCreatePeer_SchemaWithEncryptedIndex_ShouldGenerateGQL(t *testing.T) {
	test := testUtils.TestCase{
		KMS:                        testUtils.KMS{Activated: true},
		EnableSearchableEncryption: true,
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int @encryptedIndex
					}
				`,
			},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						User_encrypted(filter: {age: {_eq: 33}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"User_encrypted": gomega.Not(gomega.BeEmpty()),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestEncryptedIndexCreatePeer_AfterCreateRequest_ShouldGenerateGQL(t *testing.T) {
	test := testUtils.TestCase{
		KMS:                        testUtils.KMS{Activated: true},
		EnableSearchableEncryption: true,
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
			testUtils.CreateEncryptedIndex{
				FieldName: "age",
			},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						User_encrypted(filter: {age: {_eq: 33}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"User_encrypted": gomega.Not(gomega.BeEmpty()),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
