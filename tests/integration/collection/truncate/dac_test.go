// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package truncate

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
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
`

func TestTruncateCollectionDAC_RemovedPrivateDocumentRetainsPermissions(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   policy,
			},
			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				// Add the doc before truncate as owned by identity `1`.
				Identity: testUtils.ClientIdentity(1),
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Truncate{
				CollectionIndex: 0,
			},
			&action.AddDoc{
				CollectionID: 0,
				// Re-add the document without specifying an identity.
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Request{
				// Query the collection without an identity, no documents have been
				// returned as `John` is still owned by the identity that created it
				// before truncation.
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTruncateCollectionDAC_RemovedPublicDocumentRetainsPermissions(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   policy,
			},
			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				// Add the doc before truncate as public (no adding identity).
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Truncate{
				CollectionIndex: 0,
			},
			&action.AddDoc{
				CollectionID: 0,
				// Re-add the document using identity `1`.
				Identity: testUtils.ClientIdentity(1),
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Request{
				// Query the collection without an identity, no documents have been
				// returned as `John` is now owned by the identity that created it
				// *after* truncation, as the original, public, document was never
				// registered with ACP.
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
			&action.Request{
				// Query the document with the new identity and show that it is
				// available to the new owner.
				Identity: testUtils.ClientIdentity(1),
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
