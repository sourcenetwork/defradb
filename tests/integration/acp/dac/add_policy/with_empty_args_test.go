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

package test_acp_dac_add_policy

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_AddPolicy_EmptyPolicyData_Error(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: "",

				ExpectedError: "policy data can not be empty",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_EmptyPolicyCreator_Error(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.NoIdentity(),

				Policy: `
description: a basic policy that satisfies minimum DRI requirements
name: test
resources:
- name: users
  permissions:
  - name: delete
  - name: read
  - name: update
`,

				ExpectedError: "policy creator can not be empty",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_EmptyCreatorAndPolicyArgs_Error(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.NoIdentity(),

				Policy: "",

				ExpectedError: "policy creator can not be empty",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
