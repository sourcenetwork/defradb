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

package test_acp_nac

import (
	"testing"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_GatesDeleteDACRelationship_AuthorizedIdentity_AllowAccess(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will lose setup state when the restart happens (i.e. the restart that started nac).
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   examplePolicy,
			},
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL:      `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(3),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedExistence: false,
			},

			// This should work as the identity is authorized.
			testUtils.DeleteDACActorRelationship{
				RequestorIdentity:   testUtils.ClientIdentity(1),
				TargetIdentity:      testUtils.ClientIdentity(3),
				CollectionID:        0,
				DocID:               0,
				Relation:            "reader",
				ExpectedRecordFound: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesDeleteDACRelationship_NoIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will lose setup state when the restart happens (i.e. the restart that started nac).
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   examplePolicy,
			},
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL:      `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(3),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedExistence: false,
			},

			// We haven't authorized non-identities. So, this should error.
			testUtils.DeleteDACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(3),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedError:     testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeDeleteDACRelationPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesDeleteDACRelationship_WrongIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will lose setup state when the restart happens (i.e. the restart that started nac).
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   examplePolicy,
			},
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL:      `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(3),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedExistence: false,
			},

			// Wrong user/identity will also not be authorized.
			testUtils.DeleteDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedError:     testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeDeleteDACRelationPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
