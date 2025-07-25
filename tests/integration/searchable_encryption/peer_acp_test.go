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

const policy = `
name: Test Policy

description: A Policy

actor:
  name: actor

resources:
  users:
    permissions:
      read:
        expr: owner + reader + updater + deleter

      update:
        expr: owner + updater

      delete:
        expr: owner + deleter

      nothing:
        expr: dummy

    relations:
      owner:
        types:
          - actor

      reader:
        types:
          - actor

      updater:
        types:
          - actor

      deleter:
        types:
          - actor

      admin:
        manages:
          - reader
        types:
          - actor

      dummy:
        types:
          - actor
`

func TestDocEncryptionPeer_WithACP_ReplicatorShouldNotHaveAccess(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		SupportedDocumentACPTypes: immutable.Some(
			[]testUtils.DocumentACPType{
				testUtils.SourceHubDocumentACPType,
			},
		),
		EnableSearchableEncryption: true,
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,

				Replace: map[string]testUtils.ReplaceType{
					"Policy0": testUtils.NewPolicyIndex(0),
				},
			},
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
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": gomega.Not(gomega.Equal("John")),
							"age":  gomega.Not(gomega.Equal(21)),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
