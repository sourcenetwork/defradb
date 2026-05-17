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

// A doc registered without identity is public. _commits should return all 3
// commits (name field, age field, composite) when queried without identity.
func TestACP_QueryCommitsOnPublicDocWithoutIdentity_CanSeeCommits(t *testing.T) {
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
				Doc:          userDoc,
			},

			&action.Request{
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
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A doc registered without identity is public. _commits should return all 3
// commits when queried with any identity too.
func TestACP_QueryCommitsOnPublicDocWithIdentity_CanSeeCommits(t *testing.T) {
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
				Doc:          userDoc,
			},

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
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A doc registered with identity is private. The owner querying _commits with
// the same identity should see all 3 commits.
func TestACP_QueryCommitsOnPrivateDocWithOwnerIdentity_CanSeeCommits(t *testing.T) {
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
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A doc registered with identity is private. Querying _commits without
// identity should NOT return the doc's commits (per DAC).
func TestACP_QueryCommitsOnPrivateDocWithoutIdentity_CanNotSeeCommits(t *testing.T) {
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

// A doc registered with identity is private. Querying _commits with a
// different identity should NOT return the doc's commits (per DAC).
func TestACP_QueryCommitsOnPrivateDocWithWrongIdentity_CanNotSeeCommits(t *testing.T) {
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

// A mix of public + private doc. Querying _commits without identity should
// only return commits for the public doc (3 commits).
func TestACP_QueryCommitsOnMixedDocsWithoutIdentity_CanOnlySeePublicCommits(t *testing.T) {
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
				Doc:          userDoc,
			},

			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `
					{
						"name": "Andy",
						"age": 33
					}
				`,
			},

			&action.Request{
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
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A mix of public + private doc. Querying _commits with the owner identity
// should return commits for both docs (6 commits total).
func TestACP_QueryCommitsOnMixedDocsWithOwnerIdentity_CanSeeAllCommits(t *testing.T) {
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
				Doc:          userDoc,
			},

			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `
					{
						"name": "Andy",
						"age": 33
					}
				`,
			},

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
						{"cid": uniqueCid},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A mix of public + private doc. Querying _commits with a wrong identity
// should only return commits for the public doc (3 commits).
func TestACP_QueryCommitsOnMixedDocsWithWrongIdentity_CanOnlySeePublicCommits(t *testing.T) {
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
				Doc:          userDoc,
			},

			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `
					{
						"name": "Andy",
						"age": 33
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
