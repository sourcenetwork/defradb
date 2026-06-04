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

// Note: Eventhough this file shows we can load a policy, that has no permissions. It is important
// to know that DRI always has a set of permissions it requires. Therefore when a collection is added,
// and it has policyID and resource defined on the collection, then before we accept that collection
// the validation occurs.
// Inotherwords, we do not allow a non-DRI compliant policy to be specified on a collection, if
// it is the collection would be rejected. However we register the policy with acp even if
// the policy is not DRI compliant.

func TestACP_AddPolicy_NoPermissionsOnlyOwner_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: a policy
name: test
resources:
- name: users
`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_NoPermissionsMultiRelations_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: a policy
name: test
resources:
- name: users
  relations:
  - name: reader
    types:
    - actor
  - name: foo
  permissions:
`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_NoPermissionsLabelSingleRelation_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: a policy
name: test
resources:
- name: users
  relations:
  - name: foo
`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_NoPermissionsLabelMultiRelations_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: a policy
name: test
resources:
- name: users
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
