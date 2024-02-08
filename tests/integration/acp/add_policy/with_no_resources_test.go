// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_add_policy

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// Eventhough empty resources make no sense from a DefraDB (DPI) perspective,
// it is still a valid sourcehub policy for now.
func TestACP_AddPolicy_NoResource_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, no resource, valid policy",

		Actions: []any{
			testUtils.AddPolicy{
				IsYAML: true,

				Creator: actor1Signature,

				Policy: `
                    description: a policy

                    actor:
                      name: actor

                    resources:
                `,

				ExpectedPolicyID: "b72d8ec56ffb141922781d2b1b0803404bef57be0eeec98f1662f3017fc2de35",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Eventhough empty resources make no sense from a DefraDB (DPI) perspective,
// it is still a valid sourcehub policy for now.
func TestACP_AddPolicy_NoResourceLabel_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, no resource label, valid policy",

		Actions: []any{
			testUtils.AddPolicy{
				IsYAML: true,

				Creator: actor1Signature,

				Policy: `
                    description: a policy

                    actor:
                      name: actor
                `,

				ExpectedPolicyID: "b72d8ec56ffb141922781d2b1b0803404bef57be0eeec98f1662f3017fc2de35",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
