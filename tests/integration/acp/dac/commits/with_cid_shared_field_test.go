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

package test_acp_dac_commits

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

// Two private documents owned by different identities share an identical genesis "name"
// field block (same field value), so one block CID is owned by both documents. A
// commits(cid:) query against that shared CID must return only the commit for the document
// the requesting identity is allowed to read, never the other owner's.
func TestACP_QueryCommitsWithSharedFieldCID_ReturnsOnlyAccessibleOwner(t *testing.T) {
	test := testUtils.TestCase{
		// SignedDocs/EncryptedDocs change block cids, breaking the shared-block premise.
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

			// doc0 owned by identity 1.
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc:          `{"name": "Shared", "age": 28}`,
			},

			// doc1 owned by identity 2 - same name, so it shares doc0's "name" field block.
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(2),
				Doc:          `{"name": "Shared", "age": 29}`,
			},

			// Identity 1 can only read doc0: only doc0's owner-commit is returned.
			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `query {
						_commits(cid: "{{.FieldCID0_0_name_0}}") {
							fieldName
							docID
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"fieldName": "name", "docID": "{{.DocID0_0}}"},
					},
				},
			},

			// Identity 2 can only read doc1: only doc1's owner-commit is returned.
			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `query {
						_commits(cid: "{{.FieldCID0_1_name_0}}") {
							fieldName
							docID
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"fieldName": "name", "docID": "{{.DocID0_1}}"},
					},
				},
			},

			// No identity can read neither private document: the shared commit is hidden.
			&action.Request{
				Request: `query {
						_commits(cid: "{{.FieldCID0_0_name_0}}") {
							fieldName
							docID
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
