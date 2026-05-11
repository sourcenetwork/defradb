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
)

// The `_version` inline sub-select goes through the same dagScanNode
// code path as `_commits` (see internal/planner/select.go where DAGScan
// is invoked for the version field). The owner should see commits via
// _version against a private doc.
func TestACP_QueryVersionSubSelectOnPrivateDocWithOwnerIdentity_CanSeeCommits(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
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

			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc:          userDoc,
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `
					query {
						Users {
							_docID
							_version {
								cid
							}
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": userDocID,
							"_version": []map[string]any{
								{"cid": uniqueCid},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A wrong-identity caller cannot see the parent doc, so _version is
// never reached. Confirms the inline sub-select path is gated by the
// outer DAC fetcher and does not leak commits via a side channel.
func TestACP_QueryVersionSubSelectOnPrivateDocWithWrongIdentity_CanNotSeeCommits(t *testing.T) {
	test := testUtils.TestCase{
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

			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc:          userDoc,
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						Users {
							_docID
							_version {
								cid
							}
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
