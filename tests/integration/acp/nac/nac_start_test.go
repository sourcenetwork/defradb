// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_nac

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_StartWithDefaultConfig_NACStatusIsDisabled(t *testing.T) {
	test := testUtils.TestCase{
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
