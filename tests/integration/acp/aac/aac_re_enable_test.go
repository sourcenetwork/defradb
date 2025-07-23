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

func TestAAC_ReEnableNotConfiguredBefore_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to re-enable when aac is not configured before, return an error",
		Actions: []any{
			testUtils.ReEnableAAC{
				ExpectedError: "admin acp is not configured",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_ReEnableNotConfiguredBeforeWithIdentity_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to re-enable with identity, when aac is not configured before, return an error",
		Actions: []any{
			testUtils.ReEnableAAC{
				Identity:      testUtils.ClientIdentity(1),
				ExpectedError: "admin acp is not configured",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_ReEnableWithNoIdentityWhenTemporarilyDisabled_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to re-enable (with no identity), when aac is temporarily disabled before, return an error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},
			testUtils.DisableAAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.ReEnableAAC{
				ExpectedError: "not authorized to perform operation",
			},

			testUtils.GetAACStatus{ // Still disabled
				ExpectedStatus: client.NACDisabledTemporarily,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_ReEnableWithWrongIdentityWhenTemporarilyDisabled_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to re-enable (with wrong identity), when aac is temporarily disabled before, return an error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},
			testUtils.DisableAAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.ReEnableAAC{
				Identity:      testUtils.ClientIdentity(2),
				ExpectedError: "not authorized to perform operation",
			},

			testUtils.GetAACStatus{ // Still disabled
				ExpectedStatus: client.NACDisabledTemporarily,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_ReEnableWithValidIdentityWhenTemporarilyDisabled_AACReEnabled(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to re-enable (with valid identity), when aac is temporarily disabled before, aac re-enabled",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.DisableAAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.ReEnableAAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.GetAACStatus{ // AAC was successfully re-enabled so this won't work
				ExpectedError: "not authorized to perform operation",
			},

			testUtils.GetAACStatus{ // This works and shows that aac is enabled.
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_ReEnableWithNoIdentityWhenAlreadyEnabled_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to re-enable (with no identity), when aac is already enabled before, returns error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.ReEnableAAC{ // AAC is already enabled before.
				ExpectedError: "admin acp is already enabled",
			},

			testUtils.GetAACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_ReEnableWithWrongIdentityWhenAlreadyEnabled_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to re-enable (with wrong identity), when aac is already enabled before, returns error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.ReEnableAAC{ // AAC is already enabled before.
				Identity:      testUtils.ClientIdentity(2),
				ExpectedError: "admin acp is already enabled",
			},

			testUtils.GetAACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_ReEnableWithValidIdentityWhenAlreadyEnabled_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Try to re-enable (with valid identity), when aac is already enabled before, returns error",
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.ReEnableAAC{ // AAC is already enabled before.
				Identity:      testUtils.ClientIdentity(1),
				ExpectedError: "admin acp is already enabled",
			},

			testUtils.GetAACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_ReEnableSuccessfullyThenRestartWithNoArgs_RemainsReEnabled(t *testing.T) {
	test := testUtils.TestCase{

		SupportedDatabaseTypes: immutable.Some(
			[]testUtils.DatabaseType{
				// This test only supports file type databases since it requires the ability to
				// stop and start a node without losing data.
				testUtils.BadgerFileType,
			},
		),

		Description: "Restart when aac was re-enabled, then restarting with no args, it should remain enabled",

		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.DisableAAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.ReEnableAAC{
				Identity: testUtils.ClientIdentity(1),
			},

			// Restart with no explicit args this time, (should still start, with aac enabled).
			testUtils.Restart{},

			testUtils.GetAACStatus{ // Remains enabled even after restart.
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},

			// Can not do this as aac is enabled.
			testUtils.GetAACStatus{
				ExpectedError: "not authorized to perform operation",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_ReEnableSuccessfullyThenRestartWithStartArgs_RemainsReEnabled(t *testing.T) {
	test := testUtils.TestCase{

		SupportedDatabaseTypes: immutable.Some(
			[]testUtils.DatabaseType{
				// This test only supports file type databases since it requires the ability to
				// stop and start a node without losing data.
				testUtils.BadgerFileType,
			},
		),

		Description: "Restart when aac is enabled with args specified on start again, it should remain enabled",

		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.DisableAAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.ReEnableAAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.GetAACStatus{ // Remains enabled even after restart.
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},

			// Can not do this as aac is enabled.
			testUtils.GetAACStatus{
				ExpectedError: "not authorized to perform operation",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_ReEnableTemporarilyDisabledAACAfterRestart_ReEnabledSuccessfully(t *testing.T) {
	test := testUtils.TestCase{

		SupportedDatabaseTypes: immutable.Some(
			[]testUtils.DatabaseType{
				// This test only supports file type databases since it requires the ability to
				// stop and start a node without losing data.
				testUtils.BadgerFileType,
			},
		),

		Description: "Re-enable a temporarily disable aac after a restart, it should successfully re-enable the aac",

		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.DisableAAC{
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.Restart{},

			testUtils.GetAACStatus{ // Remains disabled ofcourse even after restart.
				ExpectedStatus: client.NACDisabledTemporarily,
			},

			testUtils.ReEnableAAC{ // Successfully re-enabled again, even after restart.
				Identity: testUtils.ClientIdentity(1),
			},

			testUtils.GetAACStatus{ // This should then not work.
				ExpectedError: "not authorized to perform operation",
			},

			// This will work and show the status.
			testUtils.GetAACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
