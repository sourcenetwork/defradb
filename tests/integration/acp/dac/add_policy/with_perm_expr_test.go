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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_AddPolicy_PermissionExprWithOwnerInTheEndWithMinus_ErrorsBecauseOwnerIsInExpr(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
name: test
description: a policy
resources:
- name: users
  permissions:
  - expr: owner
    name: delete
  - expr: reader - owner
    name: read
  - expr: owner
    name: update
  relations:
  - name: reader
    types:
    - actor
`,
				ExpectedError: "permission cannot reference `owner` relation",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_EmptyExpressionInPermission_PermissionIsAccepted(t *testing.T) {
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
  relations:
  - name: reader
    types:
    - actor
`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
