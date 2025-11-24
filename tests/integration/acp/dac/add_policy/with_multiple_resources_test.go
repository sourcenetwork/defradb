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

func TestACP_AddPolicy_MultipleResources_ValidID(t *testing.T) {
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
- name: books
  permissions:
  - expr: owner
    name: delete
  - expr: owner + reader
    name: read
  - expr: owner
    name: update
  relations:
  - name: owner
    types:
    - actor
  - name: reader
    types:
    - actor
- name: users
  permissions:
  - expr: owner
    name: delete
  - expr: owner + reader
    name: read
  - expr: owner
    name: update
  relations:
  - name: owner
    types:
    - actor
  - name: reader
    types:
    - actor
`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_MultipleResourcesUsingRelationDefinedInOther_Error(t *testing.T) {
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
- name: books
  permissions:
  - expr: owner
    name: delete
  - expr: owner + reader
    name: read
  - expr: owner
    name: update
  relations:
  - name: owner
    types:
    - actor
- name: users
  permissions:
  - expr: owner
    name: delete
  - expr: owner + reader
    name: read
  - expr: owner
    name: update
  relations:
  - name: owner
    types:
    - actor
  - name: reader
    types:
    - actor
`,

				ExpectedError: "resource books missing relation reader",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_SecondResourcesMissingRequiredOwner_Error(t *testing.T) {
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
- name: books
  permissions:
  - expr: owner
    name: delete
  - expr: owner + reader
    name: read
  - expr: owner
    name: update
  relations:
  - name: reader
    types:
    - actor
- name: users
  permissions:
  - expr: owner
    name: delete
  - expr: owner + reader
    name: read
  - expr: owner
    name: update
  relations:
  - name: owner
    types:
    - actor
  - name: reader
    types:
    - actor
`,

				ExpectedError: "BAD_INPUT",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
