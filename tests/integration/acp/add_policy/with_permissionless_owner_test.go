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

// Note: Similar to the one in ./with_no_perms_test.go
// Eventhough this file shows we can load a policy, that assigns no read/write permissions which
// are required for DPI. When a schema is loaded, and it has policyID and resource defined on the
// collection, then before we accept that schema the validation occurs. Inotherwords, we do not
// allow a non-DPI compliant policy to be specified on a collection schema, if it is, then the schema
// would be rejected. However we register the policy with acp module even if policy isn't DPI compliant.

func TestACP_AddPolicy_PermissionlessOwnerWrite_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy with owner having no write permissions, valid ID",

		Actions: []any{
			testUtils.AddPolicy{
				Creator: actor1Signature,

				Policy: `
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

				ExpectedPolicyID: "af1ee9ffe8558da8455dc1cfc5897028c16c038a053b4cf740dfcef8032d944a",
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
				Creator: actor1Signature,

				Policy: `
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

				ExpectedPolicyID: "3ceb4a4be889998496355604b68836bc280dc26dab829af3ec45b63d7767a7f1",
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
				Creator: actor1Signature,

				Policy: `
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

				ExpectedPolicyID: "af1ee9ffe8558da8455dc1cfc5897028c16c038a053b4cf740dfcef8032d944a",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
