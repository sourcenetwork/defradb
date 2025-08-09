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

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_AdminRelation_CanReEnableNAC(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will loose setup state when the restart happens (i.e. the restart that started nac).
			testUtils.DisableNAC{Identity: testUtils.ClientIdentity(1)},

			// This user, can not perform this gated operation yet.
			testUtils.ReEnableNAC{
				Identity:      testUtils.ClientIdentity(2),
				ExpectedError: "not authorized to perform operation",
			},

			// Grant access to user, but for that we need to temporarily re-enable and
			// then disable nac using the admin owner, because relationship add/delete
			// operations require admin acp to be enabled.
			testUtils.ReEnableNAC{Identity: testUtils.ClientIdentity(1)},
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2), // Grant this user "admin" relation
				Relation:          "admin",
				ExpectedExistence: false,
			},
			testUtils.DisableNAC{Identity: testUtils.ClientIdentity(1)},

			// This user, can now perform this gated operation.
			testUtils.ReEnableNAC{Identity: testUtils.ClientIdentity(2)},

			// Check if it worked, using the admin owner.
			testUtils.GetNACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
