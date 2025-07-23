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

func TestAAC_AddRelationshipWhenAACNotConfiguredBefore_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to add relationship when aac is not configured before, return an error",
		Actions: []any{
			// With requestor identity.
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedError:     "admin acp is not configured",
			},

			// Without an requestor identity.
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedError:     "admin acp is not configured",
			},

			// Without target identity.
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "admin acp is not configured",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_AddRelationshipWhenAACIsEnabledWithInvalidIdentities_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to add relationship when aac is enabled, with invalid identities, return an error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			// With wrong requestor identity (that is not authorized).
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},

			// Without an requestor identity.
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},

			// Without target identity.
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},

			// Without both identities.
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},

			// Authorized requestor identity but without target identity.
			// TODO-ACP: https://github.com/sourcenetwork/defradb/issues/3796
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "actor must be a valid did",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_AddRelationshipWhenAACIsDisabledWithInvalidIdentities_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to add relationship when aac is disabled, with invalid identities, return an error",
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
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},

			// Without an requestor identity.
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},

			// Without target identity.
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},

			// Without both identities.
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},

			// Authorized requestor identity but without target identity.
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NoIdentity(),
				Relation:          "admin",
				ExpectedError:     "operation requires ACP, but ACP not available",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_AddRelationshipWithInvalidRelationName_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Add relationship with invalid relation name, error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "unknown",
				ExpectedError:     "relation not found in resource",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_AddRelationshipWithValidIdentity_RelationshipAdded(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Add relationship with valid identity, relationship formed",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			// Should work, and not expect any to already exist.
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// Should work, as already exist (no-op).
			testUtils.AddAACActorRelationship{
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
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedExistence: false,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_AddRelationshipForAllIdentities_AllIdentitiesCanAccess(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Add relationship for all identities, all identities can acccess",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			// Should work, and not expect any to already exist.
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.AllClientIdentities(),
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// Check any normal identity can perform gated operation(s).
			testUtils.GetAACStatus{
				Identity:       testUtils.ClientIdentity(2),
				ExpectedStatus: client.NACEnabled,
			},

			// Check if an empty identity can perform gated operation(s).
			testUtils.GetAACStatus{
				Identity:       testUtils.NoIdentity(),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_AddRelationshipStillRequiresIdentityEvenIfAllIdentitiesGivenAccess_StillNeedIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Add relationship still needs identity even if all identities have access",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			// Should work, and not expect any to already exist.
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.AllClientIdentities(),
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// Check any normal identity can add a relationship.
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// But, an empty identity can not add an acp relationship.
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(4),
				Relation:          "admin",
				ExpectedError:     "admin acp relationship operation requires identity",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
