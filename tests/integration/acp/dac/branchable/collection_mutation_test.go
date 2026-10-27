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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

// Making a collection branchable does not loosen document-level access control: a stranger still
// cannot update a private document in a (branchable) permissioned collection.
func TestACP_BranchableCollection_StrangerCannotUpdateDoc(t *testing.T) {
	test := testUtils.TestCase{
		// A GQL update of an inaccessible document silently matches nothing rather than erroring, so
		// this denial is asserted against the collection update API.
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			state.CollectionNamedMutationType,
			state.CollectionSaveMutationType,
		}),
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   usersPolicy,
			},

			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL:      branchablePermissionedSDL,
			},

			// Private doc owned by identity 1.
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc:          userDoc,
			},

			// A stranger cannot update it.
			&action.UpdateDoc{
				CollectionID:  0,
				DocID:         0,
				Identity:      testUtils.ClientIdentity(2),
				Doc:           `{"age": 99}`,
				ExpectedError: "document not found or not authorized to access",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// As above for deletes: a stranger cannot delete a private document in a branchable permissioned
// collection.
func TestACP_BranchableCollection_StrangerCannotDeleteDoc(t *testing.T) {
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

			// Private doc owned by identity 1.
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc:          userDoc,
			},

			// A stranger cannot delete it.
			testUtils.DeleteDoc{
				CollectionID:  0,
				DocID:         0,
				Identity:      testUtils.ClientIdentity(2),
				ExpectedError: "document not found or not authorized to access",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Only an actor with managing access on the collection object (here, its owner) may share the
// collection-level commit DAG. A non-owner attempting to add a relationship is refused.
func TestACP_BranchableCollection_NonOwnerCannotShareCollection(t *testing.T) {
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

			// Identity 2 is not the owner of the collection object and has no managing access, so it
			// cannot grant the "reader" relation to anyone.
			&action.AddDACCollectionActorRelationship{
				CollectionID:      0,
				Relation:          "reader",
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				ExpectedError:     "failed to add document actor relationship with acp",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
