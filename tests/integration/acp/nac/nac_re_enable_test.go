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

func TestNAC_ReEnableNotConfiguredBefore_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to re-enable when nac is not configured before, return an error",
		Actions: []any{
			testUtils.ReEnableNAC{
				ExpectedError: "node acp is not configured",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_ReEnableNotConfiguredBeforeWithIdentity_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to re-enable with identity, when nac is not configured before, return an error",
		Actions: []any{
			testUtils.ReEnableNAC{
				Identity:      testUtils.ClientIdentity(1),
				ExpectedError: "node acp is not configured",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_ReEnableWithNoIdentityWhenTemporarilyDisabled_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to re-enable (with no identity), when nac is temporarily disabled before, return an error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.ReEnableNAC{
				ExpectedError: "not authorized to perform operation",
			},

			testUtils.GetNACStatus{ // Still disabled
				ExpectedStatus: client.NACDisabledTemporarily,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_ReEnableWithWrongIdentityWhenTemporarilyDisabled_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to re-enable (with wrong identity), when nac is temporarily disabled before, return an error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.ReEnableNAC{
				Identity:      testUtils.ClientIdentity(2),
				ExpectedError: "not authorized to perform operation",
			},

			testUtils.GetNACStatus{ // Still disabled
				ExpectedStatus: client.NACDisabledTemporarily,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_ReEnableWithValidIdentityWhenTemporarilyDisabled_NACReEnabled(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to re-enable (with valid identity), when nac is temporarily disabled before, nac re-enabled",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.ReEnableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.GetNACStatus{ // NAC was successfully re-enabled so this won't work
				ExpectedError: "not authorized to perform operation",
			},

			testUtils.GetNACStatus{ // This works and shows that nac is enabled.
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_ReEnableWithNoIdentityWhenAlreadyEnabled_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to re-enable (with no identity), when nac is already enabled before, returns error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.ReEnableNAC{ // NAC is already enabled before.
				ExpectedError: "node acp is already enabled",
			},

			testUtils.GetNACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_ReEnableWithWrongIdentityWhenAlreadyEnabled_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to re-enable (with wrong identity), when nac is already enabled before, returns error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.ReEnableNAC{ // NAC is already enabled before.
				Identity:      testUtils.ClientIdentity(2),
				ExpectedError: "node acp is already enabled",
			},

			testUtils.GetNACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_ReEnableWithValidIdentityWhenAlreadyEnabled_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to re-enable (with valid identity), when nac is already enabled before, returns error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.ReEnableNAC{ // NAC is already enabled before.
				Identity:      testUtils.ClientIdentity(1),
				ExpectedError: "node acp is already enabled",
			},

			testUtils.GetNACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_ReEnableSuccessfullyThenRestartWithNoArgs_RemainsReEnabled(t *testing.T) {
	test := testUtils.TestCase{

		SupportedDatabaseTypes: immutable.Some(
			[]state.DatabaseType{
				// This test only supports file type databases since it requires the ability to
				// stop and start a node without losing data.
				testUtils.BadgerFileType,
			},
		),

		Description: "Restart when nac was re-enabled, then restarting with no args, it should remain enabled",

		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.ReEnableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			// Restart with no explicit args this time, (should still start, with nac enabled).
			testUtils.Restart{},

			testUtils.GetNACStatus{ // Remains enabled even after restart.
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},

			// Can not do this as nac is enabled.
			testUtils.GetNACStatus{
				ExpectedError: "not authorized to perform operation",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_ReEnableSuccessfullyThenRestartWithStartArgs_RemainsReEnabled(t *testing.T) {
	test := testUtils.TestCase{

		SupportedDatabaseTypes: immutable.Some(
			[]state.DatabaseType{
				// This test only supports file type databases since it requires the ability to
				// stop and start a node without losing data.
				testUtils.BadgerFileType,
			},
		),

		Description: "Restart when nac is enabled with args specified on start again, it should remain enabled",

		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.ReEnableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.GetNACStatus{ // Remains enabled even after restart.
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},

			// Can not do this as nac is enabled.
			testUtils.GetNACStatus{
				ExpectedError: "not authorized to perform operation",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_ReEnableTemporarilyDisabledNACAfterRestart_ReEnabledSuccessfully(t *testing.T) {
	test := testUtils.TestCase{

		SupportedDatabaseTypes: immutable.Some(
			[]state.DatabaseType{
				// This test only supports file type databases since it requires the ability to
				// stop and start a node without losing data.
				testUtils.BadgerFileType,
			},
		),

		Description: "Re-enable a temporarily disable nac after a restart, it should successfully re-enable the nac",

		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.Restart{},

			testUtils.GetNACStatus{ // Remains disabled ofcourse even after restart.
				ExpectedStatus: client.NACDisabledTemporarily,
			},

			testUtils.ReEnableNAC{ // Successfully re-enabled again, even after restart.
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.GetNACStatus{ // This should then not work.
				ExpectedError: "not authorized to perform operation",
			},

			// This will work and show the status.
			testUtils.GetNACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
