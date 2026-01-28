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

	"github.com/onsi/gomega"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

const policy = `
name: Test Policy

description: A Policy

resources:
- name: users
  permissions:
  - name: read
    expr: reader + updater + deleter
  - name: update
    expr: updater
  - name: delete
    expr: deleter
  - name: nothing
    doc: |
      placeholder permission, to show a policy can contain any user defined relation,
      in addition to the defra required ones
    expr: dummy

  relations:
  - name: reader
    types:
    - actor

  - name: updater
    types:
    - actor

  - name: deleter
    types:
    - actor

  - name: admin
    manages:
    - reader
    types:
    - actor

  - name: dummy
    types:
    - actor
`

func TestDocEncryptionPeer_WithACP_ReplicatorShouldNotHaveAccess(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		SupportedDocumentACPTypes: immutable.Some(
			[]state.DocumentACPType{
				state.LocalDocumentACPType,
			},
		),
		EnableSearchableEncryption: true,
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.AddDACPolicy{
				Identity: testUtils.NodeIdentity(0),
				Policy:   policy,
			},
			&action.AddSchema{
				Schema: `
					type User @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int @encryptedIndex
					}
				`,
			},
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.CreateDoc{
				Identity: testUtils.NodeIdentity(0),
				NodeID:   immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
				IsDocEncrypted: true,
			},
			testUtils.WaitForSESync{},
			&action.Request{
				NodeID:   immutable.Some(0),
				Identity: testUtils.NodeIdentity(0),
				Request: `
					query {
						encrypted_User(filter: {age: {_eq: 21}}) {
							docIDs
						}
					}
				`,
				Results: map[string]any{
					"encrypted_User": []map[string]any{
						{
							"docIDs": gomega.ConsistOf(testUtils.DocIDAt(0, 0)),
						},
					},
				},
			},
			&action.Request{
				NodeID:   immutable.Some(1),
				Identity: testUtils.NodeIdentity(0),
				Request: `
					query {
						User {
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
			&action.Request{
				NodeID:   immutable.Some(1),
				Identity: testUtils.NodeIdentity(0),
				Request: `
					query {
						_commits {
							delta
						}
					}
				`,
				// this replicator doesn't have access to the document, so it doesn't have the related
				// commit blocks. Once we introduce a dedicated permission for replication,
				// this should be updated to return the commits with encrypted deltas.
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
