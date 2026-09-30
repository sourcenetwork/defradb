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

package test_acp_dac_query

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

const usersPolicy = `
description: a test policy
name: test
resources:
- name: users
  permissions:
  - name: delete
  - name: read
    expr: reader
  - name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: reader
    types:
    - actor
`

// Two private documents owned by different identities share an identical genesis "name" field
// block, so its CID is owned by both documents. A time-travel query against that shared CID
// returns each owning document's state at that version, but only for the documents the
// requesting identity is allowed to read.
func TestACP_QuerySimpleWithSharedFieldCid_ReturnsOnlyAccessibleOwners(t *testing.T) {
	test := testUtils.TestCase{
		MultiplierExcludes: []string{multiplier.SignedDocs, multiplier.EncryptedDocs},
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   usersPolicy,
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
			// doc 0 owned by identity 1, doc 1 owned by identity 2 - same name, so they share
			// the genesis "name" field block.
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc:          `{"name": "Shared", "age": 28}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(2),
				Doc:          `{"name": "Shared", "age": 29}`,
			},
			// Identity 1 can only read doc 0.
			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `query {
					Users (cid: "{{.FieldCID0_0_name_0}}") {
						_docID
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"_docID": testUtils.NewDocIndex(0, 0), "name": "Shared"},
					},
				},
			},
			// Identity 2 can only read doc 1.
			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `query {
					Users (cid: "{{.FieldCID0_1_name_0}}") {
						_docID
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"_docID": testUtils.NewDocIndex(0, 1), "name": "Shared"},
					},
				},
			},
			// No identity can read neither private document.
			&action.Request{
				Request: `query {
					Users (cid: "{{.FieldCID0_0_name_0}}") {
						_docID
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
