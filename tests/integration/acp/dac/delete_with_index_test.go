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

package test_acp_dac

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// TestACP_DAC_DeleteDocByID_WithReaderOnlyAccess_DoesNotOrphanIndex verifies that when an
// identity with read-but-not-delete access attempts to delete a document by ID, the request
// is denied and the document's index entries remain intact.
func TestACP_DAC_DeleteDocByID_WithReaderOnlyAccess_DoesNotOrphanIndex(t *testing.T) {
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
  - expr: reader + deleter
    name: read
  - expr: deleter
    name: delete
  - expr: updater
    name: update
  relations:
  - name: deleter
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
				Doc: `{
					"name": "John",
					"age": 30
				}`,
			},

			// Give identity 2 the reader relation only
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedExistence: false,
			},

			// Identity 2 can read the document.
			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `query {
					Users {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "John", "age": int64(30)},
					},
				},
			},

			// Identity 2 attempts to delete by ID, but should be denied.
			testUtils.DeleteDoc{
				Identity:      testUtils.ClientIdentity(2),
				CollectionID:  0,
				DocID:         0,
				ExpectedError: "document not found or not authorized to access",
			},

			// After the failed delete, the owner must still find the document via the
			// index. If index entries had been stripped by the unauthorized attempt, this
			// filter-based query (which uses the index) would return nothing.
			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `query {
					Users(filter: {name: {_eq: "John"}}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "John", "age": int64(30)},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
