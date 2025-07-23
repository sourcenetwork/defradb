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

func TestAAC_RestartNodeWithAACEnabledWithoutAACArgs_RestartsAndAACIsStillEnabled(t *testing.T) {
	test := testUtils.TestCase{
		SupportedDatabaseTypes: immutable.Some(
			[]testUtils.DatabaseType{
				// This test only supports file type databases since it requires the ability to
				// stop and start a node without losing data.
				testUtils.BadgerFileType,
			},
		),

		Description: "Restart node after it has aac started before (without aac args), aac should still be enabled.",

		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			// Restart with no explicit args this time, (should still start).
			testUtils.Restart{},

			testUtils.GetAACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_RestartNodeWithAACEnabledWithExplicitlySpecifyingSameArgs_RestartsAndAACIsStillEnabled(t *testing.T) {
	test := testUtils.TestCase{
		SupportedDatabaseTypes: immutable.Some(
			[]testUtils.DatabaseType{
				// This test only supports file type databases since it requires the ability to
				// stop and start a node without losing data.
				testUtils.BadgerFileType,
			},
		),

		Description: "Restart node with same args again, after it has aac started before, aac should still be enabled.",

		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			// Restart, but given same identity explicitly as before, (should still start).
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			testUtils.GetAACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_RestartNodeWithAACEnabledWithAnotherIdentity_IgnoreNewIdentityAndRestartWithExistingAACState(t *testing.T) {
	test := testUtils.TestCase{
		SupportedDatabaseTypes: immutable.Some(
			[]testUtils.DatabaseType{
				// This test only supports file type databases since it requires the ability to
				// stop and start a node without losing data.
				testUtils.BadgerFileType,
			},
		),

		Description: "Restart node with new identity arg, after it has aac started before, recover existing aac state.",

		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			// Restart and recover AAC state that was already configured before, ignore this new identity.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(2),
				EnableAAC: true,
			},

			testUtils.GetAACStatus{
				Identity:       testUtils.ClientIdentity(1),
				ExpectedStatus: client.NACEnabled,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
