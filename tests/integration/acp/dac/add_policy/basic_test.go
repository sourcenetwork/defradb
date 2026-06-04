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

func TestACP_AddPolicy_BasicYAML_ValidPolicyID(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
