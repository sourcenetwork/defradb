// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_aac

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestAAC_StartWithDefaultConfig_AACStatusIsDisabled(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Start node with default config, aac status is disabled",
		Actions: []any{
			testUtils.GetAACStatus{
				ExpectedStatus: client.NACNotConfigured,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_StartWithDefaultConfigWithIdentity_AACStatusIsDisabled(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Start node with default config, aac status is disabled even with an Identity",
		Actions: []any{
			testUtils.GetAACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACNotConfigured,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_StartNodeWithIdentityAndWithAACEnableTrue_AACEnabledSuccessfully(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Start node with an identity and --aac-enable=true, enable aac.",
		Actions: []any{
			testUtils.GetAACStatus{
				ExpectedStatus: client.NACNotConfigured,
			},

			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.GetAACStatus{ // Now we need valid identity to perform this operation.
				ExpectedError: client.ErrNotAuthorizedToPerformOperation.Error(),
			},

			testUtils.GetAACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_StartNodeNoIdentityWithAACEnableTrue_ErrorAsIdentityIsNeeded(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Start node with no identity with --aac-enable=true, error as identity must be provided.",
		Actions: []any{
			testUtils.GetAACStatus{
				ExpectedStatus: client.NACNotConfigured,
			},

			testUtils.Close{},
			testUtils.Start{
				Identity:      testUtils.NoIdentity(),
				EnableAAC:     true,
				ExpectedError: client.ErrCanNotStartAACWithoutIdentity.Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_StartNodeWithIdentityAndWithAACEnableFalse_AACNotEnabled(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Start node with an identity and --aac-enable=false, aac does not start.",
		Actions: []any{
			testUtils.GetAACStatus{
				ExpectedStatus: client.NACNotConfigured,
			},

			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: false,
			},

			testUtils.GetAACStatus{
				ExpectedStatus: client.NACNotConfigured,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
