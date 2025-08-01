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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_AdminRelation_CanDeleteNACRelationship(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Grant admin relation, gain ability to add nac relationship",
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will loose setup state when the restart happens (i.e. the restart that started nac).
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(3), // Try deleting relationship for this actor.
				Relation:          "admin",
				ExpectedExistence: false,
			},
			// Note: Setup to test relationship deletion with is now done.

			// This user, can not perform this gated operation yet.
			testUtils.DeleteNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				Relation:          "admin",
				ExpectedError:     "not authorized to perform operation",
			},

			// Grant access to user.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2), // Grant this user "admin" relation
				Relation:          "admin",
				ExpectedExistence: false,
			},

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
