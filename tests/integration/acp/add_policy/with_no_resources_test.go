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
				Identity: actor1Identity,

				Policy: `
                    description: a policy

                    actor:
                      name: actor

                    resources:
                `,

				ExpectedPolicyID: "e16824022121b55f2b2babbd2ab82960a8837767197e20acf9c577cbb4539991",
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
				Identity: actor1Identity,

				Policy: `
                    description: a policy

                    actor:
                      name: actor
                `,

				ExpectedPolicyID: "e16824022121b55f2b2babbd2ab82960a8837767197e20acf9c577cbb4539991",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Eventhough empty resources make no sense from a DefraDB (DPI) perspective,
// it is still a valid sourcehub policy for now.
func TestACP_AddPolicy_PolicyWithOnlySpace_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, adding a policy that has only space",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: " ",

				ExpectedPolicyID: "e16824022121b55f2b2babbd2ab82960a8837767197e20acf9c577cbb4539991",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
