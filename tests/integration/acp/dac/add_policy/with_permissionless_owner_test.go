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

// Note: Similar to the one in ./with_no_perms_test.go
// Eventhough this file shows we can load a policy, that assigns no read/update/delete permissions which
// are required for DRI. When a schema is loaded, and it has policyID and resource defined on the
// collection, then before we accept that schema the validation occurs. Inotherwords, we do not
// allow a non-DRI compliant policy to be specified on a collection schema, if it is, then the schema
// would be rejected. However we register the policy with acp even if policy isn't DRI compliant.

func TestACP_AddPolicy_PermissionlessOwnerUpdate_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy with owner having no update permissions, valid ID",

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: test
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          update:
                            expr: reader
                          delete:
                            expr: owner
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_PermissionlessOwnerDelete_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy with owner having no delete permissions, valid ID",

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: test
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          update:
                            expr: owner
                          delete:
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_PermissionlessOwnerRead_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy with owner having no read permissions, valid ID",

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: test
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          update:
                            expr: owner + reader
                          delete:
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_PermissionlessOwnerReadUpdateDelete_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy with owner having no read/update/delete permissions, valid ID",

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: test
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          update:
                            expr: reader
                          delete:
                            expr: reader
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
