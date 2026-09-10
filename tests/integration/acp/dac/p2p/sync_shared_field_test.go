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

package test_acp_dac_p2p

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
	"github.com/sourcenetwork/defradb/tests/state"
)

const sharedFieldSyncPolicy = `
description: A Policy
name: Test Policy
resources:
- name: users
  permissions:
  - expr: deleter
    name: delete
  - expr: dummy
    name: nothing
  - expr: reader + updater + deleter
    name: read
  - expr: updater
    name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: deleter
    types:
    - actor
  - name: dummy
    types:
    - actor
  - name: reader
    types:
    - actor
  - name: updater
    types:
    - actor
`

// sharedFieldSyncTestCase builds a two-node DAC scenario where node 0 holds two private
// documents that share an identical genesis "name" field block (same field value => one block
// CID owned by both documents). Node 1 is granted reader access to exactly one of them (owned
// by identity 1) and must sync that document in full - including the shared "name" field
// block, which is also owned by the second document (owned by identity 2) that node 1 cannot
// read.
//
// The access check that gates sending the shared block resolves the block's owners from the
// owner index (returned in lexicographic docID order) and must grant access when ANY owner is
// readable, not just the first. The granted document's age differs between the two callers, so
// across the two tests the accessible owner sorts both first and second within the shared
// block's owner list - exercising the access check regardless of lexicographic order.
func sharedFieldSyncTestCase(grantedAge, otherAge int) testUtils.TestCase {
	return testUtils.TestCase{
		SupportedDocumentACPTypes: immutable.Some(
			[]state.DocumentACPType{
				state.LocalDocumentACPType,
				state.RemoteDocumentACPType,
			},
		),
		// Signing/encryption change block cids per document, which would break the
		// shared-block premise this test relies on.
		//
		// This test withholds one document from node 1 on purpose, so one of the
		// heads never arrives. An external node is polled for its commits instead
		// of read from its event bus, and that poll cannot express a head that is
		// not coming.
		// https://github.com/sourcenetwork/defradb/issues/5193
		MultiplierExcludes: []string{
			multiplier.SignedDocs,
			multiplier.EncryptedDocs,
			multiplier.CrossVersionOldSource,
			multiplier.CrossVersionNewSource,
		},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),

			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},

			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   sharedFieldSyncPolicy,
			},

			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			// The document node 1 will be granted access to (owned by identity 1).
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "Shared",
					"age":  grantedAge,
				},
			},
			// A second document with the same "name", so it shares the genesis "name" field
			// block. It is owned by a different identity and never granted to node 1, so node 1
			// must never receive it - only the shared block it co-owns.
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(2),
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "Shared",
					"age":  otherAge,
				},
			},

			// Grant node 1 reader access to the first (identity 1) document only. The
			// relationship is set on node 0 (the sender), whose access check decides which
			// blocks to send.
			testUtils.AddDACActorRelationship{
				NodeID:            immutable.Some(0),
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NodeIdentity(1),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedExistence: false,
			},

			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},

			testUtils.SyncDocs{
				NodeID:       1,
				CollectionID: 0,
				DocIDs:       []int{0},
				SourceNodes:  []int{0},
			},

			testUtils.WaitForSync{},

			// Node 1 synced the granted document, and its shared "name" field block synced too:
			// "name" is present, not just "age". If the access check only consulted the first
			// owner of the shared block, the field block would have been withheld whenever the
			// other document sorted first, leaving "name" empty.
			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				NodeID:   immutable.Some(1),
				Request: `query {
						Users {
							name
							age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Shared", "age": int64(grantedAge)},
					},
				},
			},
		},
	}
}

// Node 1 is granted access to the document whose "age" is the smaller value.
func TestACP_P2PSyncSharedFieldBlock_AccessToFirstDoc(t *testing.T) {
	testUtils.ExecuteTestCase(t, sharedFieldSyncTestCase(10, 20))
}

// Node 1 is granted access to the document whose "age" is the larger value. Together with the
// test above this exercises the access check regardless of the lexicographic order of the
// shared block's owners.
func TestACP_P2PSyncSharedFieldBlock_AccessToSecondDoc(t *testing.T) {
	testUtils.ExecuteTestCase(t, sharedFieldSyncTestCase(20, 10))
}
