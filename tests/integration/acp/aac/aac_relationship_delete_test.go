// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_aac

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestAAC_DeleteRelationshipWhenAACNotConfiguredBefore_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to delete relationship when aac is not configured before, return an error",
		Actions: []any{
			// With requestor identity.
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedError:     "admin acp is not configured",
			},

			// Without an requestor identity.
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedError:     "admin acp is not configured",
			},

			// Without target identity.
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "admin acp is not configured",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_DeleteRelationshipWhenAACIsEnabledWithInvalidIdentities_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to delete relationship when aac is enabled, with invalid identities, return an error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			// With wrong requestor identity (that is not authorized).
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},

			// Without an requestor identity.
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},

			// Without target identity.
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},

			// Without both identities.
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},

			// Authorized requestor identity but without target identity.
			// TODO-ACP: https://github.com/sourcenetwork/defradb/issues/3796
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity:   testUtils.ClientIdentity(1),
				TargetIdentity:      testUtils.NoIdentity(),
				Relation:            "admin",
				ExpectedRecordFound: false,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_DeleteRelationshipWhenAACIsDisabledWithInvalidIdentities_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to delete relationship when aac is disabled, with invalid identities, return an error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},
			testUtils.DisableAAC{
				Identity: testUtils.ClientIdentity(1),
			},

			// With wrong requestor identity (that is not authorized).
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},

			// Without an requestor identity.
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},

			// Without target identity.
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},

			// Without both identities.
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},

			// Authorized requestor identity but without target identity.
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_DeleteRelationshipWithInvalidRelationName_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Delete relationship with invalid relation name, error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.DeleteAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "unknown",
				ExpectedError:     "relation not found in resource",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_DeleteRelationshipWithValidIdentity_RelationshipDeleted(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Delete relationship with valid identity, relationship deleted",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},
			// This is just setup to test deletion works.
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// Should work, and not expect any to already exist.
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity:   testUtils.ClientIdentity(1),
				TargetIdentity:      testUtils.ClientIdentity(2),
				Relation:            "admin",
				ExpectedRecordFound: true,
			},

			// Should work, as already exist (no-op).
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity:   testUtils.ClientIdentity(1),
				TargetIdentity:      testUtils.ClientIdentity(2),
				Relation:            "admin",
				ExpectedRecordFound: false,
			},

			// Check this identity can now not do gated operation(s), as access has been revoked.
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_DeleteRelationshipForAllIdentities_AllImplicitIdentitiesAccessRevoked(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Delete relationship for * identities, acccess revoked from * identities",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},
			// This is just setup to test deletion works.
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(3), // Explicitly allow this identity before.
				Relation:          "admin",
				ExpectedExistence: false,
			},
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.AllClientIdentities(), // Implicitly allow all identities.
				Relation:          "admin",
				ExpectedExistence: false,
			},
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(4), // Explicitly allow this identity after.
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// Should work, and ofcourse find a record that was deleted.
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity:   testUtils.ClientIdentity(1),
				TargetIdentity:      testUtils.AllClientIdentities(), // Re-voke from all implicitly allowed identities
				Relation:            "admin",
				ExpectedRecordFound: true,
			},

			// Check any normal identity is no longer allowed to perform gated operation(s).
			testUtils.GetAACStatus{
				Identity:      testUtils.ClientIdentity(2),
				ExpectedError: "not authorized to perform operation",
			},

			// Check if an empty identity can no longer perform gated operation(s).
			testUtils.GetAACStatus{
				Identity:      testUtils.NoIdentity(),
				ExpectedError: "not authorized to perform operation",
			},

			// Check that explicitly allowed identities still have access to perform gated operation(s).
			testUtils.GetAACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
			testUtils.GetAACStatus{
				Identity:       testUtils.ClientIdentity(3),
				ExpectedStatus: client.NACEnabled,
			},
			testUtils.GetAACStatus{
				Identity:       testUtils.ClientIdentity(4),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_DeleteRelationshipStillRequiresIdentityEvenIfAllIdentitiesGivenAccess_StillNeedIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Delete relationship still needs identity even if all identities have access",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},
			// This is just setup to test deletion works.
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.AllClientIdentities(),
				Relation:          "admin",
				ExpectedExistence: false,
			},
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// An empty identity can not delete an acp relationship.
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "admin acp relationship operation requires identity",
			},

			// Any normal identity can delete a relationship.
			testUtils.DeleteAACActorRelationship{
				RequestorIdentity:   testUtils.ClientIdentity(2),
				TargetIdentity:      testUtils.ClientIdentity(3),
				Relation:            "admin",
				ExpectedRecordFound: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
