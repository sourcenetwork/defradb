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

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_GatesAddDACRelationship_AuthorizedIdentity_AllowAccess(t *testing.T) {
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

			// This should work as the identity is authorized.
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(3),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedExistence: false,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesAddDACRelationship_NoIdentity_NotAuthorizedError(t *testing.T) {
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

			// We haven't authorized non-identities. So, this should error.
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(3),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedError:     testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeAddDACRelationPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesAddDACRelationship_WrongIdentity_NotAuthorizedError(t *testing.T) {
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

			// Wrong user/identity will also not be authorized.
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedError:     testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeAddDACRelationPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
