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

package test_acp_nac

import (
	"testing"

	"github.com/sourcenetwork/lens/host-go/config/model"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestNAC_GatesAddLens_AuthorizedIdentity_AllowAccess(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			// This should work as the identity is authorized.
			&action.AddLens{
				Identity: testUtils.ClientIdentity(1),
				Lens: model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.SetDefaultModulePath,
							Arguments: map[string]any{
								"dst":   "name",
								"value": "John",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesAddLens_NoIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			// We haven't authorized non-identities. So, this should error.
			&action.AddLens{
				Identity: testUtils.NoIdentity(),
				Lens: model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.SetDefaultModulePath,
							Arguments: map[string]any{
								"dst":   "name",
								"value": "John",
							},
						},
					},
				},
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeAddLensPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesAddLens_WrongIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			// Wrong user/identity will also not be authorized.
			&action.AddLens{
				Identity: testUtils.ClientIdentity(2),
				Lens: model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.SetDefaultModulePath,
							Arguments: map[string]any{
								"dst":   "name",
								"value": "John",
							},
						},
					},
				},
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeAddLensPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
