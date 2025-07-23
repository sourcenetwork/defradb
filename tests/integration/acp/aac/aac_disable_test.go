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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestAAC_DisableNotConfiguredBefore_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to disable when aac is not configured before, return an error",
		Actions: []any{
			testUtils.DisableAAC{
				ExpectedError: "admin acp is not configured",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_DisableNotConfiguredBeforeWithIdentity_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to disable with identity, when aac is not configured before, return an error",
		Actions: []any{
			testUtils.DisableAAC{
				Identity:      testUtils.ClientIdentity(1),
				ExpectedError: "admin acp is not configured",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_DisableWithoutIdentityOnNodeThatHasConfigured_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to disable without identity, when aac is configured and started before, return an error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.DisableAAC{
				ExpectedError: "not authorized to perform operation",
			},

			testUtils.GetAACStatus{ // Did not disable.
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_DisableWithWrongIdentityOnNodeThatHasConfigured_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to disable with wrong identity, when aac is configured and started before, return an error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.DisableAAC{
				Identity:      testUtils.ClientIdentity(2),
				ExpectedError: "not authorized to perform operation",
			},

			testUtils.GetAACStatus{ // Did not disable.
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_DisableWithIdentityOnNodeThatHasAACConfiguredAndEnabled_Successful(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to disable with identity, when aac is configured and started before, successful",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			// Can not do this request without identity before disabling.
			testUtils.GetAACStatus{
				ExpectedError: "not authorized to perform operation",
			},

			testUtils.DisableAAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.GetAACStatus{ // Did disable.
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACDisabledTemporarily,
			},

			// Can do the status request without identity now also.
			testUtils.GetAACStatus{
				ExpectedStatus: client.NACDisabledTemporarily,
			},

			// Can do the status request with any identity now also.
			testUtils.GetAACStatus{
				Identity:       testUtils.ClientIdentity(2),
				ExpectedStatus: client.NACDisabledTemporarily,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_DisableNoIdentityWhenConfiguredAndAlreadyDisabledBefore_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to disable (no identity), when aac is configured but temporarily disabled already, return error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.DisableAAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.DisableAAC{ // Try to disable again, without identity.
				ExpectedError: "admin acp is already disabled",
			},

			testUtils.GetAACStatus{ // Still remain disabled and accesible without identity.
				ExpectedStatus: client.NACDisabledTemporarily,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_DisableWithIdentityWhenConfiguredAndAlreadyDisabledBefore_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to disable (with identity), when aac is configured but temporarily disabled already, return error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.DisableAAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.DisableAAC{ // Try to disable again, with identity.
				Identity:      testUtils.ClientIdentity(1),
				ExpectedError: "admin acp is already disabled",
			},

			testUtils.GetAACStatus{ // Still remain disabled and accesible without identity.
				ExpectedStatus: client.NACDisabledTemporarily,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_DisableSuccessfullyThenRestartWithNoArgs_RemainsDisabled(t *testing.T) {
	test := testUtils.TestCase{

		SupportedDatabaseTypes: immutable.Some(
			[]testUtils.DatabaseType{
				// This test only supports file type databases since it requires the ability to
				// stop and start a node without losing data.
				testUtils.BadgerFileType,
			},
		),

		Description: "Restart when aac is temporarily disabled, then restarting with no args it should remain disabled",

		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.DisableAAC{
				Identity: testUtils.ClientIdentity(1),
			},

			// Restart with no explicit args this time, (should still start).
			testUtils.Restart{},

			testUtils.GetAACStatus{ // Remains disabled even after restart.
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACDisabledTemporarily,
			},

			// Can stil do the status request without identity.
			testUtils.GetAACStatus{
				ExpectedStatus: client.NACDisabledTemporarily,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_DisableSuccessfullyThenRestartWithStartArgs_RemainsDisabled(t *testing.T) {
	test := testUtils.TestCase{

		SupportedDatabaseTypes: immutable.Some(
			[]testUtils.DatabaseType{
				// This test only supports file type databases since it requires the ability to
				// stop and start a node without losing data.
				testUtils.BadgerFileType,
			},
		),

		Description: "Restart when aac is temporarily disabled with args specified on start again, it should remain disabled",

		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.DisableAAC{
				Identity: testUtils.ClientIdentity(1),
			},

			// Note: This is because the start command only configures the aac system for the first time,
			// tying the aac system to the identity owner. After that to disable and re-enable the user
			// must use the specific aac client commands.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.GetAACStatus{ // Remains disabled even after restart.
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACDisabledTemporarily,
			},

			// Can stil do the status request without identity.
			testUtils.GetAACStatus{
				ExpectedStatus: client.NACDisabledTemporarily,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
