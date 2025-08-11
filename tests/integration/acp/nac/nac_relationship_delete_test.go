// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_nac

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_DeleteRelationshipWhenNACNotConfiguredBefore_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// With requestor identity.
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedError:     "node acp is not configured",
			},

			// Without an requestor identity.
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedError:     "node acp is not configured",
			},

			// Without target identity.
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "node acp is not configured",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_DeleteRelationshipWhenNACIsEnabledWithInvalidIdentities_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			// With wrong requestor identity (that is not authorized).
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},

			// Without an requestor identity.
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},

			// Without target identity.
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},

			// Without both identities.
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},

			// Authorized requestor identity but without target identity.
			// TODO-ACP: https://github.com/sourcenetwork/defradb/issues/3796
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity:   testUtils.ClientIdentity(1),
				TargetIdentity:      testUtils.NoIdentity(),
				Relation:            "admin",
				ExpectedRecordFound: false,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_DeleteRelationshipWhenNACIsDisabledWithInvalidIdentities_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			// With wrong requestor identity (that is not authorized).
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},

			// Without an requestor identity.
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},

			// Without target identity.
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},

			// Without both identities.
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},

			// Authorized requestor identity but without target identity.
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_DeleteRelationshipWithInvalidRelationName_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.DeleteNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "unknown",
				ExpectedError:     "relation not found in resource",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_DeleteRelationshipWithValidIdentity_RelationshipDeleted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// This is just setup to test deletion works.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// Should work, and not expect any to already exist.
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity:   testUtils.ClientIdentity(1),
				TargetIdentity:      testUtils.ClientIdentity(2),
				Relation:            "admin",
				ExpectedRecordFound: true,
			},

			// Should work, as already exist (no-op).
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity:   testUtils.ClientIdentity(1),
				TargetIdentity:      testUtils.ClientIdentity(2),
				Relation:            "admin",
				ExpectedRecordFound: false,
			},

			// Check this identity can now not do gated operation(s), as access has been revoked.
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_DeleteRelationshipForAllIdentities_AllImplicitIdentitiesAccessRevoked(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// This is just setup to test deletion works.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(3), // Explicitly allow this identity before.
				Relation:          "admin",
				ExpectedExistence: false,
			},
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.AllClientIdentities(), // Implicitly allow all identities.
				Relation:          "admin",
				ExpectedExistence: false,
			},
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(4), // Explicitly allow this identity after.
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// Should work, and ofcourse find a record that was deleted.
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity:   testUtils.ClientIdentity(1),
				TargetIdentity:      testUtils.AllClientIdentities(), // Re-voke from all implicitly allowed identities
				Relation:            "admin",
				ExpectedRecordFound: true,
			},

			// Check any normal identity is no longer allowed to perform gated operation(s).
			testUtils.GetNACStatus{
				Identity:      testUtils.ClientIdentity(2),
				ExpectedError: "not authorized to perform operation",
			},

			// Check if an empty identity can no longer perform gated operation(s).
			testUtils.GetNACStatus{
				Identity:      testUtils.NoIdentity(),
				ExpectedError: "not authorized to perform operation",
			},

			// Check that explicitly allowed identities still have access to perform gated operation(s).
			testUtils.GetNACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
			testUtils.GetNACStatus{
				Identity:       testUtils.ClientIdentity(3),
				ExpectedStatus: client.NACEnabled,
			},
			testUtils.GetNACStatus{
				Identity:       testUtils.ClientIdentity(4),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_DeleteRelationshipStillRequiresIdentityEvenIfAllIdentitiesGivenAccess_StillNeedIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// This is just setup to test deletion works.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.AllClientIdentities(),
				Relation:          "admin",
				ExpectedExistence: false,
			},
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// An empty identity can not delete an acp relationship.
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "node acp relationship operation requires identity",
			},

			// Any normal identity can delete a relationship.
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity:   testUtils.ClientIdentity(2),
				TargetIdentity:      testUtils.ClientIdentity(3),
				Relation:            "admin",
				ExpectedRecordFound: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
