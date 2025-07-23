// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_aac_relation_admin

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestAAC_AdminRelation_CanReEnableAAC(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Grant admin relation, gain ability to re-enable aac",
		Actions: []any{
			// Starting with ACC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},
			// Note: Doing setup steps after starting with aac enabled, otherwise the in-memory tests
			// will loose setup state when the restart happens (i.e. the restart that started aac).
			testUtils.DisableAAC{Identity: testUtils.ClientIdentity(1)},

			// This user, can not perform this gated operation yet.
			testUtils.ReEnableAAC{
				Identity:      testUtils.ClientIdentity(2),
				ExpectedError: "not authorized to perform operation",
			},

			// Grant access to user, but for that we need to temporarily re-enable and
			// then disable aac using the admin owner, because relationship add/delete
			// operations require admin acp to be enabled.
			testUtils.ReEnableAAC{Identity: testUtils.ClientIdentity(1)},
			testUtils.AddAACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2), // Grant this user "admin" relation
				Relation:          "admin",
				ExpectedExistence: false,
			},
			testUtils.DisableAAC{Identity: testUtils.ClientIdentity(1)},

			// This user, can now perform this gated operation.
			testUtils.ReEnableAAC{Identity: testUtils.ClientIdentity(2)},

			// Check if it worked, using the admin owner.
			testUtils.GetAACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
