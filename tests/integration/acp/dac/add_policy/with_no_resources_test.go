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
	"github.com/sourcenetwork/defradb/tests/state"
)

// Eventhough empty resources make no sense from a DefraDB (DRI) perspective,
// it is still a valid sourcehub policy for now.
func TestACP_AddPolicy_NoResource_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: test
                    description: a policy
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

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: test
                    description: a policy
                `,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A Policy can have no resources (incompatible with DRI) but it needs a name.
func TestACP_AddPolicy_PolicyWithOnlySpace_NameIsRequired(t *testing.T) {
	test := testUtils.TestCase{
		SupportedDocumentACPTypes: immutable.Some([]state.DocumentACPType{
			// This is currently a local-acp only limitation, this test-restriction
			// can be lifted if/when SourceHub introduces the same limitation.
			state.LocalDocumentACPType,
		}),
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: " ",

				ExpectedError: "name required",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
