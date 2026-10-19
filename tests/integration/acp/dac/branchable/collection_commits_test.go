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
// When the collection is created with an identity it is registered as a (private) acp object. The
// owner can read every commit of the collection's DAG.
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

// A private branchable collection gates its ENTIRE commit DAG, not just the collection-level commit.
// Even though the document here is PUBLIC (created without identity), a stranger cannot see any of
// its commits, because every commit of a branchable collection is also gated on the (private)
// collection object.
func TestACP_BranchableCollectionCommits_StrangerCannotSeeAnyCommit(t *testing.T) {
	ownerCid := testUtils.NewUniqueValue()

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

			// Public doc - but the collection it belongs to is private.
			&action.AddDoc{
				CollectionID: 0,
				Doc:          userDoc,
			},

			// Owner sees all 4 commits.
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

			// Stranger sees nothing - the whole DAG is gated on the private collection object.
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

// A request without any identity is treated like a stranger: it cannot see any commit of a private
// branchable collection.
func TestACP_BranchableCollectionCommits_NoIdentityCannotSeeAnyCommit(t *testing.T) {
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
					"_commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// After the owner grants a stranger the "reader" relation on the collection object, the stranger can
// read the whole DAG of the (otherwise private) branchable collection: here the doc is public, so
// collection-level read access is all that is needed to see every commit.
func TestACP_BranchableCollectionCommits_SharedCollectionReaderCanSeeAllCommits(t *testing.T) {
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

			// Public doc - visibility is therefore gated only by the collection object.
			&action.AddDoc{
				CollectionID: 0,
				Doc:          userDoc,
			},

			// Before sharing: stranger sees nothing.
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

			// Owner shares read access to the collection's commit DAG with the stranger.
			&action.AddDACCollectionActorRelationship{
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

// Access is additive: a stranger granted ONLY the collection "reader" relation can see the
// collection-level commit, but the document commits remain gated on the (private) document object,
// which the stranger has no access to. So the stranger sees just the single collection-level commit.
func TestACP_BranchableCollectionCommits_CollectionReaderStillGatedByPrivateDoc(t *testing.T) {
	collectionCid := testUtils.NewUniqueValue()

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

			// Private doc - owned by identity 1.
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc:          userDoc,
			},

			// Share only the collection object (not the document) with the stranger.
			&action.AddDACCollectionActorRelationship{
				CollectionID:      0,
				Relation:          "reader",
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				ExpectedExistence: false,
			},

			// Stranger sees only the collection-level commit; the 3 document commits remain gated
			// on the private document object.
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
						{"cid": collectionCid},
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
