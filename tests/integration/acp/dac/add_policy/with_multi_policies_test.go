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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_AddPolicy_AddDuplicatePolicyByOtherCreator_ValidPolicyIDs(t *testing.T) {
	const policyUsedByBoth string = `
description: a policy
name: test
resources:
- name: users
  permissions:
  - name: delete
  - name: read
  - name: update
`

	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: policyUsedByBoth,

				ExpectedPolicyID: immutable.Some(
					"1239a04400966b311339f62db50044b1bde70cece2ce9897d69c1bafa5cfab81",
				),
			},

			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(2),

				Policy: policyUsedByBoth,

				ExpectedPolicyID: immutable.Some(
					"166758b3b8f5edd06f46e9079b30b701aadcf3e59e64afe1c46d4242924bd850",
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
description: a policy
name: test
resources:
- name: users
  permissions:
  - name: delete
  - name: read
  - name: update
`,

				ExpectedPolicyID: immutable.Some(
					"1239a04400966b311339f62db50044b1bde70cece2ce9897d69c1bafa5cfab81",
				),
			},

			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: a policy
name: test
resources:
- name: users
  permissions:
  - name: delete
  - name: read
  - name: update
`,

				ExpectedPolicyID: immutable.Some(
					"166758b3b8f5edd06f46e9079b30b701aadcf3e59e64afe1c46d4242924bd850",
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
description: a policy
name: test
resources:
- name: users
  permissions:
  - name: delete
  - name: read
  - name: update
`,

				ExpectedPolicyID: immutable.Some(
					"1239a04400966b311339f62db50044b1bde70cece2ce9897d69c1bafa5cfab81",
				),
			},

			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: a policy
name: test
resources:
- name: users
  permissions:
  - name: delete
  - name: read
  - name: update
`,

				ExpectedPolicyID: immutable.Some(
					"166758b3b8f5edd06f46e9079b30b701aadcf3e59e64afe1c46d4242924bd850",
				),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_AddMultipleDifferentPolicies_ValidPolicyIDs(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: a policy
                    description: a policy
                    resources:
                    - name: users
                      permissions:
                      - name: read
                      - name: update
                      - name: delete
                `,
			},

			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: a policy
                    description: another policy
                    resources:
                    - name: users
                      permissions:
                      - name: read
                        expr: reader
                      - name: update
                      - name: delete
                      relations:
                      - name: reader
                        types:
                        - actor
                      - name: admin
                        manages:
                        - reader
                        types:
                        - actor
                `,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
