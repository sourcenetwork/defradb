// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_dac_add_policy

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_AddPolicy_AddDuplicatePolicyByOtherCreator_ValidPolicyIDs(t *testing.T) {
	const policyUsedByBoth string = `
actor:
  name: actor
description: a policy
name: test
resources:
- name: users
  permissions:
  - expr: owner
    name: delete
  - expr: owner
    name: read
  - expr: owner
    name: update
  relations:
  - name: owner
    types:
    - actor
`

	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: policyUsedByBoth,

				ExpectedPolicyID: immutable.Some(
					"60079fa5b415dfc6f6e6b70e123a8acb8de26d94d7ff9410449fb12950963ff0",
				),
			},

			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(2),

				Policy: policyUsedByBoth,

				ExpectedPolicyID: immutable.Some(
					"4f113ea28e09992fdf6f3a8ccac8be8d8d39c932f48f54c42fff9c3513cd9a7a",
				),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_AddMultipleDuplicatePolicies_Error(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
actor:
  name: actor
description: a policy
name: test
resources:
- name: users
  permissions:
  - expr: owner
    name: delete
  - expr: owner
    name: read
  - expr: owner
    name: update
  relations:
  - name: owner
    types:
    - actor
`,

				ExpectedPolicyID: immutable.Some(
					"60079fa5b415dfc6f6e6b70e123a8acb8de26d94d7ff9410449fb12950963ff0",
				),
			},

			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
actor:
  name: actor
description: a policy
name: test
resources:
- name: users
  permissions:
  - expr: owner
    name: delete
  - expr: owner
    name: read
  - expr: owner
    name: update
  relations:
  - name: owner
    types:
    - actor
`,

				ExpectedPolicyID: immutable.Some(
					"4f113ea28e09992fdf6f3a8ccac8be8d8d39c932f48f54c42fff9c3513cd9a7a",
				),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_AddMultipleDuplicatePoliciesDifferentFmts_ProducesDifferentIDs(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
actor:
  name: actor
description: a policy
name: test
resources:
- name: users
  permissions:
  - expr: owner
    name: delete
  - expr: owner
    name: read
  - expr: owner
    name: update
  relations:
  - name: owner
    types:
    - actor
`,

				ExpectedPolicyID: immutable.Some(
					"60079fa5b415dfc6f6e6b70e123a8acb8de26d94d7ff9410449fb12950963ff0",
				),
			},

			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
actor:
  name: actor
description: a policy
name: test
resources:
- name: users
  permissions:
  - expr: owner
    name: delete
  - expr: owner
    name: read
  - expr: owner
    name: update
  relations:
  - name: owner
    types:
    - actor
`,

				ExpectedPolicyID: immutable.Some(
					"4f113ea28e09992fdf6f3a8ccac8be8d8d39c932f48f54c42fff9c3513cd9a7a",
				),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
