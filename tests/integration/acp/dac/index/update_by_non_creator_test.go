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

package test_acp_dac_index

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// Single-node regression guard: a non-creator with the "updater"
// relation can update an indexed field without corrupting the index.
// The P2P sibling lives in tests/integration/acp/dac/p2p/.
//
// The policy grants "updater" both update and read (via "read: reader
// + updater + deleter") so the non-creator can read the prior doc
// version that updateIndexedDoc needs to delete the old index entry.
func TestACPWithIndex_NonCreatorUpdatesIndexedField_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
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
    - updater
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
						name: String @index
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc: `
					{
						"name": "Alice",
						"age": 30
					}
				`,
			},
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				CollectionID:      0,
				DocID:             0,
				Relation:          "updater",
				ExpectedExistence: false,
			},
			&action.UpdateDoc{
				Identity:     testUtils.ClientIdentity(2),
				CollectionID: 0,
				DocID:        0,
				Doc: `
					{
						"name": "Alice Updated"
					}
				`,
			},
			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `
					query {
						Users {
							name
							age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Alice Updated", "age": int64(30)},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
