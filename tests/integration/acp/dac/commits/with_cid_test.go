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

// Private doc. _commits(cid:) without identity should NOT return the
// targeted commit (per DAC).
func TestACP_QueryCommitsWithCIDOnPrivateDocWithoutIdentity_CanNotSeeCommit(t *testing.T) {
	test := testUtils.TestCase{
		// Result CIDs are hardcoded; SignedDocs mode changes cid values.
		MultiplierExcludes: []string{multiplier.SignedDocs},
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
				Request: `
					query {
						_commits(cid: "` + userDocCompositeCid + `") {
							cid
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Private doc. _commits(cid:) with the owner identity should return the
// targeted commit.
func TestACP_QueryCommitsWithCIDOnPrivateDocWithOwnerIdentity_CanSeeCommit(t *testing.T) {
	test := testUtils.TestCase{
		MultiplierExcludes: []string{multiplier.SignedDocs},
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
						_commits(cid: "` + userDocCompositeCid + `") {
							cid
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"cid": userDocCompositeCid},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
