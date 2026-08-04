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
)

func TestPurgeRemovesSearchableEncryptionArtifacts(t *testing.T) {
	test := testUtils.TestCase{
		KMS:                        testUtils.KMS{Activated: true},
		EnableSearchableEncryption: true,
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `type User {
					name: String @encryptedIndex
				}`,
			},
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.AddDoc{
				NodeID:         immutable.Some(0),
				CollectionID:   0,
				Doc:            `{"name":"alice"}`,
				IsDocEncrypted: true,
			},
			testUtils.WaitForSync{},
			testUtils.WaitForSESync{},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					encrypted_User(filter: {name: {_eq: "alice"}}) {
						docIDs
					}
				}`,
				Results: map[string]any{
					"encrypted_User": []map[string]any{{
						"docIDs": gomega.ConsistOf(testUtils.DocIDAt(0, 0)),
					}},
				},
			},
			&action.PurgeDocs{
				NodeID:          immutable.Some(1),
				CollectionIndex: 0,
				DocIndexes:      []int{0},
			},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					encrypted_User(filter: {name: {_eq: "alice"}}) {
						docIDs
					}
				}`,
				Results: map[string]any{
					"encrypted_User": []map[string]any{{
						"docIDs": gomega.BeEmpty(),
					}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
