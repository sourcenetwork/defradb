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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// Note: Similar to the one in ./with_no_perms_test.go
// Eventhough this file shows we can load a policy, that assigns no read/write permissions which
// are required for DPI. When a schema is loaded, and it has policyID and resource defined on the
// collection, then before we accept that schema the validation occurs. Inotherwords, we do not
// allow a non-DPI compliant policy to be specified on a collection schema, if it is, then the schema
// would be rejected. However we register the policy with acp even if policy isn't DPI compliant.

func TestACP_AddPolicy_PermissionlessOwnerWrite_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy with owner having no write permissions, valid ID",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: immutable.Some(1),

				Policy: `
                    name: test
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          write:
                            expr: reader
                          read:
                            expr: owner + reader

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                `,

				ExpectedPolicyID: "9328e41c1969c6269bfd82162b45831ccec8df9fc8d57902620ad43baaa0d77d",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_PermissionlessOwnerRead_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy with owner having no read permissions, valid ID",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: immutable.Some(1),

				Policy: `
                    name: test
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          write:
                            expr: owner + reader
                          read:
                            expr: reader

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                `,

				ExpectedPolicyID: "74f3c0996d5b1669b9efda5ef45f93a925df9f770e2dcd53f352b5f0693a2b0f",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_PermissionlessOwnerReadWrite_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy with owner having no read/write permissions, valid ID",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: immutable.Some(1),

				Policy: `
                    name: test
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          write:
                            expr: reader
                          read:
                            expr: owner + reader

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                `,

				ExpectedPolicyID: "9328e41c1969c6269bfd82162b45831ccec8df9fc8d57902620ad43baaa0d77d",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
