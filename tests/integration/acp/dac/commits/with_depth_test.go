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

// Private doc + an update. _commits(depth: 1) with the owner identity should
// return only the latest head commits (composite + name head + age head = 3).
func TestACP_QueryCommitsWithDepthOnPrivateDocWithOwnerIdentity_CanSeeCommits(t *testing.T) {
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

			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `
					{
						"age": 29
					}
				`,
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `
					query {
						_commits(depth: 1) {
							cid
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"cid": uniqueCid},
						{"cid": uniqueCid},
						{"cid": uniqueCid},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Private doc + an update. _commits(depth: 1) with a wrong identity should
// NOT see any commits (per DAC).
func TestACP_QueryCommitsWithDepthOnPrivateDocWithWrongIdentity_CanNotSeeCommits(t *testing.T) {
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

			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `
					{
						"age": 29
					}
				`,
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						_commits(depth: 1) {
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
