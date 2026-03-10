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

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

func TestNAC_GatesTruncateCollection_AuthorizedIdentity_AllowAccess(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will lose setup state when the restart happens (i.e. the restart that started nac).
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users {}
				`,
			},
			// This should work as the identity is authorized.
			&action.Truncate{
				Identity:        testUtils.ClientIdentity(1),
				CollectionIndex: 0,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesTruncateCollection_NoIdentity_NotAuthorizedError(t *testing.T) {
	// todo: Investigate and test this behavior across all client types when implementing granular NAC permissions.
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
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users {}
				`,
			},
			// We haven't authorized non-identities. So, this should error.
			&action.Truncate{
				Identity:        testUtils.NoIdentity(),
				CollectionIndex: 0,
				ExpectedError:   testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeTruncateCollectionPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesTruncateCollection_NoIdentity_CLIandCandHTTPClient_NotAuthorizedError(t *testing.T) {
	// todo: Investigate and test this behavior across all client types when implementing granular NAC permissions.
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
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users {}
				`,
			},
			// We haven't authorized non-identities. So, this should error.
			&action.Truncate{
				Identity:        testUtils.NoIdentity(),
				CollectionIndex: 0,
				ExpectedError:   testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeGetCollectionPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesTruncateCollection_WrongIdentity_NotAuthorizedError(t *testing.T) {
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
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users {}
				`,
			},
			// Wrong user/identity will also not be authorized.
			&action.Truncate{
				Identity:        testUtils.ClientIdentity(2),
				CollectionIndex: 0,
				ExpectedError:   testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeTruncateCollectionPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesTruncateCollection_WrongIdentity_CLIandHTTPClient_NotAuthorizedError(t *testing.T) {
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
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users {}
				`,
			},
			// Wrong user/identity will also not be authorized.
			&action.Truncate{
				Identity:        testUtils.ClientIdentity(2),
				CollectionIndex: 0,
				ExpectedError:   testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeGetCollectionPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
