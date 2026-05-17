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

// A `_commits` query without docID iterates heads across every collection
// in the database. The owner should see all commits from both private
// collections; the per-versionID cache must route each block to the
// correct collection for the access check.
func TestACP_QueryCommitsMultiCollectionWithOwnerIdentity_CanSeeAllCommits(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   usersAndPostsPolicy,
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

					type Posts @policy(
						id: "{{.Policy0}}",
						resource: "posts"
					) {
						title: String
					}
				`,
			},

			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc:          userDoc,
			},

			&action.AddDoc{
				CollectionID: 1,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `
					{
						"title": "First Post"
					}
				`,
			},

			// 3 commits per Users doc (name, age, composite) + 2 commits
			// per Posts doc (title, composite) = 5 commits.
			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `
					query {
						_commits {
							cid
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"cid": uniqueCid},
						{"cid": uniqueCid},
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

// A wrong-identity caller should see no commits across either collection,
// confirming the per-versionID cache routing does not leak between
// collections.
func TestACP_QueryCommitsMultiCollectionWithWrongIdentity_CanNotSeeCommits(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   usersAndPostsPolicy,
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

					type Posts @policy(
						id: "{{.Policy0}}",
						resource: "posts"
					) {
						title: String
					}
				`,
			},

			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc:          userDoc,
			},

			&action.AddDoc{
				CollectionID: 1,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `
					{
						"title": "First Post"
					}
				`,
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						_commits {
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
