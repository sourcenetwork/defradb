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

func TestNAC_AddRelationshipWhenNACNotConfiguredBefore_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to add relationship when nac is not configured before, return an error",
		Actions: []any{
			// With requestor identity.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedError:     "node acp is not configured",
			},

			// Without an requestor identity.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedError:     "node acp is not configured",
			},

			// Without target identity.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "node acp is not configured",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_AddRelationshipWhenNACIsEnabledWithInvalidIdentities_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to add relationship when nac is enabled, with invalid identities, return an error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			// With wrong requestor identity (that is not authorized).
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},

			// Without an requestor identity.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},

			// Without target identity.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},

			// Without both identities.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},

			// Authorized requestor identity but without target identity.
			// TODO-ACP: https://github.com/sourcenetwork/defradb/issues/3796
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "actor must be a valid did",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_AddRelationshipWhenNACIsDisabledWithInvalidIdentities_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to add relationship when nac is disabled, with invalid identities, return an error",
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
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},

			// Without an requestor identity.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},

			// Without target identity.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},

			// Without both identities.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},

			// Authorized requestor identity but without target identity.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_AddRelationshipWithInvalidRelationName_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Add relationship with invalid relation name, error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "unknown",
				ExpectedError:     "relation not found in resource",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_AddRelationshipWithValidIdentity_RelationshipAdded(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Add relationship with valid identity, relationship formed",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			// Should work, and not expect any to already exist.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// Should work, as already exist (no-op).
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedExistence: true,
			},

			// Check if this identity can now do gated operation(s), to make someone else admin.
			// Note: This is possible because in the policy `admin` manages `admin` relation:
			// `relations:
			//   admin:
			//     manages:
			//       - admin
			//     types:
			//       - actor
			//`
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedExistence: false,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_AddRelationshipForAllIdentities_AllIdentitiesCanAccess(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Add relationship for all identities, all identities can acccess",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			// Should work, and not expect any to already exist.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.AllClientIdentities(),
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// Check any normal identity can perform gated operation(s).
			testUtils.GetNACStatus{
				Identity:       testUtils.ClientIdentity(2),
				ExpectedStatus: client.NACEnabled,
			},

			// Check if an empty identity can perform gated operation(s).
			testUtils.GetNACStatus{
				Identity:       testUtils.NoIdentity(),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_AddRelationshipStillRequiresIdentityEvenIfAllIdentitiesGivenAccess_StillNeedIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Add relationship still needs identity even if all identities have access",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			// Should work, and not expect any to already exist.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.AllClientIdentities(),
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// Check any normal identity can add a relationship.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// But, an empty identity can not add an acp relationship.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(4),
				Relation:          "admin",
				ExpectedError:     "node acp relationship operation requires identity",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
