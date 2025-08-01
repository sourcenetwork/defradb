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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestNAC_DisableNotConfiguredBefore_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to disable when nac is not configured before, return an error",
		Actions: []any{
			testUtils.DisableNAC{
				ExpectedError: "node acp is not configured",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_DisableNotConfiguredBeforeWithIdentity_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to disable with identity, when nac is not configured before, return an error",
		Actions: []any{
			testUtils.DisableNAC{
				Identity:      testUtils.ClientIdentity(1),
				ExpectedError: "node acp is not configured",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_DisableWithoutIdentityOnNodeThatHasConfigured_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to disable without identity, when nac is configured and started before, return an error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.DisableNAC{
				ExpectedError: "not authorized to perform operation",
			},

			testUtils.GetNACStatus{ // Did not disable.
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_DisableWithWrongIdentityOnNodeThatHasConfigured_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to disable with wrong identity, when nac is configured and started before, return an error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.DisableNAC{
				Identity:      testUtils.ClientIdentity(2),
				ExpectedError: "not authorized to perform operation",
			},

			testUtils.GetNACStatus{ // Did not disable.
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_DisableWithIdentityOnNodeThatHasNACConfiguredAndEnabled_Successful(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to disable with identity, when nac is configured and started before, successful",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			// Can not do this request without identity before disabling.
			testUtils.GetNACStatus{
				ExpectedError: "not authorized to perform operation",
			},

			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.GetNACStatus{ // Did disable.
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACDisabledTemporarily,
			},

			// Can do the status request without identity now also.
			testUtils.GetNACStatus{
				ExpectedStatus: client.NACDisabledTemporarily,
			},

			// Can do the status request with any identity now also.
			testUtils.GetNACStatus{
				Identity:       testUtils.ClientIdentity(2),
				ExpectedStatus: client.NACDisabledTemporarily,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_DisableNoIdentityWhenConfiguredAndAlreadyDisabledBefore_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to disable (no identity), when nac is configured but temporarily disabled already, return error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.DisableNAC{ // Try to disable again, without identity.
				ExpectedError: "node acp is already disabled",
			},

			testUtils.GetNACStatus{ // Still remain disabled and accesible without identity.
				ExpectedStatus: client.NACDisabledTemporarily,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_DisableWithIdentityWhenConfiguredAndAlreadyDisabledBefore_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to disable (with identity), when nac is configured but temporarily disabled already, return error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.DisableNAC{ // Try to disable again, with identity.
				Identity:      testUtils.ClientIdentity(1),
				ExpectedError: "node acp is already disabled",
			},

			testUtils.GetNACStatus{ // Still remain disabled and accesible without identity.
				ExpectedStatus: client.NACDisabledTemporarily,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_DisableSuccessfullyThenRestartWithNoArgs_RemainsDisabled(t *testing.T) {
	test := testUtils.TestCase{

		SupportedDatabaseTypes: immutable.Some(
			[]state.DatabaseType{
				// This test only supports file type databases since it requires the ability to
				// stop and start a node without losing data.
				testUtils.BadgerFileType,
			},
		),

		Description: "Restart when nac is temporarily disabled, then restarting with no args it should remain disabled",

		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			// Restart with no explicit args this time, (should still start).
			testUtils.Restart{},

			testUtils.GetNACStatus{ // Remains disabled even after restart.
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACDisabledTemporarily,
			},

			// Can stil do the status request without identity.
			testUtils.GetNACStatus{
				ExpectedStatus: client.NACDisabledTemporarily,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_DisableSuccessfullyThenRestartWithStartArgs_RemainsDisabled(t *testing.T) {
	test := testUtils.TestCase{

		SupportedDatabaseTypes: immutable.Some(
			[]state.DatabaseType{
				// This test only supports file type databases since it requires the ability to
				// stop and start a node without losing data.
				testUtils.BadgerFileType,
			},
		),

		Description: "Restart when nac is temporarily disabled with args specified on start again, it should remain disabled",

		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			// Note: This is because the start command only configures the nac system for the first time,
			// tying the nac system to the identity owner. After that to disable and re-enable the user
			// must use the specific nac client commands.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.GetNACStatus{ // Remains disabled even after restart.
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACDisabledTemporarily,
			},

			// Can stil do the status request without identity.
			testUtils.GetNACStatus{
				ExpectedStatus: client.NACDisabledTemporarily,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
