// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_nac_relation_admin

import (
	"testing"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_AdminRelation_WithDACManagerRelation_CanAddDACActorRelationship(t *testing.T) {
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
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema:   `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				CollectionID:      0,
				DocID:             0,
				// Note: If the other user is not a manager it can't perform relationship operations
				Relation: "manager",
			},

			// This user, can not perform this gated operation yet, eventhough they are a DAC manager.
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedError:     testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeDACRelationAddPerm),
			},

			// Grant access to user.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2), // Grant this user "admin" relation
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// This user, can now perform this gated operation.
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
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

func TestNAC_AdminRelation_WithoutManagerDACRelation_CanNotAddDACActorRelationship(t *testing.T) {
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
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema:   `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				CollectionID:      0,
				DocID:             0,
				// Note: This is a reader and not a manger so can't perform DAC relationship operations.
				Relation: "reader",
			},

			// This user, can not perform this gated operation yet, eventhough they are a DAC manager.
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedError:     testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeDACRelationAddPerm),
			},

			// Grant access to user.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2), // Grant this user "admin" relation
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// This user, can now try to perform this gated operation.
			// But, since this user is not a DAC manager, they can't add a DAC actor relationship.
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedError:     "cannot create relationship: actor is not a manager of relation",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
