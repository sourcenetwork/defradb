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

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestNAC_GatesIndexDrop_AuthorizedIdentity_AllowAccess(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.GoClientType,
				state.HTTPClientType,
				state.CLIClientType,
				state.CClientType,
			},
		),
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will lose setup state when the restart happens (i.e. the restart that started nac).
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type User {
						name: String @index
					}
				`,
			},

			// This should work as the identity is authorized.
			&action.DropIndex{
				Identity:  testUtils.ClientIdentity(1),
				IndexName: "User_name_ASC",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesIndexDrop_NoIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		// todo: Investigate and test this behavior across all client types when implementing granular NAC permissions.
		// See: https://github.com/sourcenetwork/defradb/issues/4383
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.GoClientType,
				state.JSClientType,
			},
		),
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will lose setup state when the restart happens (i.e. the restart that started nac).
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type User {
						name: String @index
					}
				`,
			},

			// We haven't authorized non-identities. So, this should error.
			&action.DropIndex{
				Identity:      testUtils.NoIdentity(),
				IndexName:     "User_name_ASC",
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeIndexDropPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesIndexDrop_NoIdentity_CLIandCandHTTPClient_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		// todo: Investigate and test this behavior across all client types when implementing granular NAC permissions.
		// See: https://github.com/sourcenetwork/defradb/issues/4383
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.CClientType,
				state.HTTPClientType,
				state.CLIClientType,
			},
		),
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will lose setup state when the restart happens (i.e. the restart that started nac).
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type User {
						name: String @index
					}
				`,
			},

			// We haven't authorized non-identities. So, this should error.
			&action.DropIndex{
				Identity:      testUtils.NoIdentity(),
				IndexName:     "User_name_ASC",
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeCollectionGetPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesIndexDrop_WrongIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		// todo: Investigate and test this behavior across all client types when implementing granular NAC permissions.
		// See: https://github.com/sourcenetwork/defradb/issues/4383
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.GoClientType,
				state.JSClientType,
			},
		),
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will lose setup state when the restart happens (i.e. the restart that started nac).
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type User {
						name: String @index
					}
				`,
			},

			// Wrong user/identity will also not be authorized.
			&action.DropIndex{
				Identity:      testUtils.ClientIdentity(2),
				IndexName:     "User_name_ASC",
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeIndexDropPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesIndexDrop_WrongIdentity_CLIandCandHTTPClient_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		// todo: Investigate and test this behavior across all client types when implementing granular NAC permissions.
		// See: https://github.com/sourcenetwork/defradb/issues/4383
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.CClientType,
				state.CLIClientType,
				state.HTTPClientType,
			},
		),
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will lose setup state when the restart happens (i.e. the restart that started nac).
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type User {
						name: String @index
					}
				`,
			},

			// Wrong user/identity will also not be authorized.
			&action.DropIndex{
				Identity:      testUtils.ClientIdentity(2),
				IndexName:     "User_name_ASC",
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeCollectionGetPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
