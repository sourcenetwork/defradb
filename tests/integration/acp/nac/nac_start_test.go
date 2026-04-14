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

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_StartWithDefaultConfig_NACStatusIsDisabled(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Starting a node with default config results in NAC being disabled.",
		Actions: []any{
			testUtils.GetNACStatus{
				ExpectedStatus: client.NACNotConfigured,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_StartWithDefaultConfigWithIdentity_NACStatusIsDisabled(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Starting a node with default config and an identity still results in NAC being disabled.",
		Actions: []any{
			testUtils.GetNACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACNotConfigured,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_StartNodeWithIdentityAndWithNACEnableTrue_NACEnabledSuccessfully(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Starting a node with an identity and NAC enable flag set to true enables NAC successfully.",
		Actions: []any{
			testUtils.GetNACStatus{
				ExpectedStatus: client.NACNotConfigured,
			},

			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.GetNACStatus{ // Now we need valid identity to perform this operation.
				ExpectedError: client.ErrNotAuthorizedToPerformOperation.Error(),
			},

			testUtils.GetNACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_StartNodeNoIdentityWithNACEnableTrue_ErrorAsIdentityIsNeeded(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Starting a node with NAC enable true but no identity returns an error.",
		Actions: []any{
			testUtils.GetNACStatus{
				ExpectedStatus: client.NACNotConfigured,
			},

			testUtils.Close{},
			testUtils.Start{
				Identity:      testUtils.NoIdentity(),
				EnableNAC:     true,
				ExpectedError: client.ErrCanNotStartNACWithoutIdentity.Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_StartNodeWithIdentityAndWithNACEnableFalse_NACNotEnabled(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Starting a node with an identity but NAC enable false does not enable NAC.",
		Actions: []any{
			testUtils.GetNACStatus{
				ExpectedStatus: client.NACNotConfigured,
			},

			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: false,
			},

			testUtils.GetNACStatus{
				ExpectedStatus: client.NACNotConfigured,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
