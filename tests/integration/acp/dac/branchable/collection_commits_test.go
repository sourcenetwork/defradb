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

package test_acp_dac_branchable

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// A branchable collection with one 2-field doc produces 4 commits: the name field commit, the age
// field commit, the composite commit, and the collection-level commit.
//
// When the collection is created with an identity it is registered as a (private) acp object, and
// when the doc is created with the same identity the doc is private too. The owner sees all 4.
func TestACP_BranchableCollectionCommits_OwnerCanSeeAllCommits(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   usersPolicy,
			},

			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL:      branchablePermissionedSDL,
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
						{"cid": uniqueCid},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This isolates the feature: the doc is PUBLIC (created without identity) but the collection object
// is PRIVATE (collection created with identity 1). A stranger can see the 3 public doc commits but
// NOT the collection-level commit, which is gated on the (private) collection object.
//
// Without branchable-collection acp the stranger would see all 4 commits (the collection-level
// commit would leak), so this asserts exactly 3.
func TestACP_BranchableCollectionCommits_StrangerCannotSeeCollectionCommit(t *testing.T) {
	// A fresh matcher is needed per request, as the uniqueness matcher tracks seen values across
	// every request within a test run.
	ownerCid := testUtils.NewUniqueValue()
	strangerCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   usersPolicy,
			},

			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL:      branchablePermissionedSDL,
			},

			// Public doc - its commits are visible to everyone.
			&action.AddDoc{
				CollectionID: 0,
				Doc:          userDoc,
			},

			// Owner sees all 4 (3 doc + 1 collection-level).
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
						{"cid": ownerCid},
						{"cid": ownerCid},
						{"cid": ownerCid},
						{"cid": ownerCid},
					},
				},
				NonOrderedResults: true,
			},

			// Stranger sees only the 3 public doc commits, the collection-level commit is gated.
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
						{"cid": strangerCid},
						{"cid": strangerCid},
						{"cid": strangerCid},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A request without any identity is treated like a stranger: it can see the public doc commits but
// not the private collection-level commit.
func TestACP_BranchableCollectionCommits_NoIdentityCannotSeeCollectionCommit(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   usersPolicy,
			},

			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL:      branchablePermissionedSDL,
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

// After the owner grants a stranger the "reader" relation on the collection object, the stranger
// can see the collection-level commit too (4 commits instead of 3).
func TestACP_BranchableCollectionCommits_SharedReaderCanSeeCollectionCommit(t *testing.T) {
	// A fresh matcher is needed per request, as the uniqueness matcher tracks seen values across
	// every request within a test run.
	beforeCid := testUtils.NewUniqueValue()
	afterCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   usersPolicy,
			},

			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL:      branchablePermissionedSDL,
			},

			// Public doc so the stranger can already see the 3 doc commits, isolating the effect
			// of sharing the collection object.
			&action.AddDoc{
				CollectionID: 0,
				Doc:          userDoc,
			},

			// Before sharing: stranger sees only the 3 public doc commits.
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
						{"cid": beforeCid},
						{"cid": beforeCid},
						{"cid": beforeCid},
					},
				},
				NonOrderedResults: true,
			},

			// Owner shares read access to the collection's commit DAG with the stranger.
			testUtils.AddDACCollectionActorRelationship{
				CollectionID:      0,
				Relation:          "reader",
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				ExpectedExistence: false,
			},

			// After sharing: stranger sees all 4 commits.
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
						{"cid": afterCid},
						{"cid": afterCid},
						{"cid": afterCid},
						{"cid": afterCid},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A branchable, permissioned collection created WITHOUT an identity is not registered as an acp
// object, so its collection-level commit DAG is public (consistent with public documents). A
// stranger can see all 4 commits.
func TestACP_BranchableCollectionCommits_PublicCollectionIsVisibleToAll(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   usersPolicy,
			},

			// No identity -> the collection object is not registered -> public commit DAG.
			&action.AddCollection{
				SDL: branchablePermissionedSDL,
			},

			&action.AddDoc{
				CollectionID: 0,
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
					"_commits": []map[string]any{
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
