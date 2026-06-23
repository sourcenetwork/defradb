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
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestACP_P2PUpdatePrivateDocumentsOnDifferentNodes_SourceHubACP(t *testing.T) {
	test := testUtils.TestCase{

		SupportedDocumentACPTypes: immutable.Some(
			[]state.DocumentACPType{
				state.SourceHubDocumentACPType,
			},
		),

		Actions: []any{
			testUtils.RandomNetworkingConfig(),

			testUtils.RandomNetworkingConfig(),

			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
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
`,
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

			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},

			&action.AddDoc{
				Identity: testUtils.ClientIdentity(1),

				NodeID: immutable.Some(0),

				CollectionID: 0,

				DocMap: map[string]any{
					"name": "Shahzad",
				},
			},

			&action.AddDoc{
				Identity: testUtils.ClientIdentity(1),

				NodeID: immutable.Some(1),

				CollectionID: 0,

				DocMap: map[string]any{
					"name": "Shahzad Lone",
				},
			},

			testUtils.WaitForSync{},

			&action.UpdateDoc{
				Identity: testUtils.ClientIdentity(1),

				NodeID: immutable.Some(0),

				CollectionID: 0,

				DocID: 0,

				Doc: `
					{
						"name": "ShahzadLone"
					}
				`,
			},

			&action.UpdateDoc{
				Identity: testUtils.ClientIdentity(1),

				NodeID: immutable.Some(1),

				CollectionID: 0,

				DocID: 1,

				Doc: `
					{
						"name": "ShahzadLone"
					}
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
