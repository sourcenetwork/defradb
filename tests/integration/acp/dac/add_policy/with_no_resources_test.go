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

// Eventhough empty resources make no sense from a DefraDB (DRI) perspective,
// it is still a valid sourcehub policy for now.
func TestACP_AddPolicy_NoResource_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, no resource, valid policy",

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: test
                    description: a policy

                    actor:
                      name: actor

                    resources:
                `,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Eventhough empty resources make no sense from a DefraDB (DRI) perspective,
// it is still a valid sourcehub policy for now.
func TestACP_AddPolicy_NoResourceLabel_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, no resource label, valid policy",

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: test
                    description: a policy

                    actor:
                      name: actor
                `,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A Policy can have no resources (incompatible with DRI) but it needs a name.
func TestACP_AddPolicy_PolicyWithOnlySpace_NameIsRequired(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, adding a policy that has only space",
		SupportedDocumentACPTypes: immutable.Some([]testUtils.DocumentACPType{
			// This is currently a local-acp only limitation, this test-restriction
			// can be lifted if/when SourceHub introduces the same limitation.
			testUtils.LocalDocumentACPType,
		}),
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: " ",

				ExpectedError: "name is required",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
