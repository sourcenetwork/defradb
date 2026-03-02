// Copyright 2026 Democratized Data Foundation
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

func TestNAC_GatesDeleteEncryptedIndex_AuthorizedIdentity_AllowAccess(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.HTTPClientType,
				state.CLIClientType,
				state.GoClientType,
				state.CClientType,
			},
		),
		Actions: []any{
			// Starting all nodes with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users {
						name: String
					}
				`,
			},

			testUtils.AddEncryptedIndex{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				FieldName:    "name",
			},

			// This should work as the identity is authorized.
			testUtils.DeleteEncryptedIndex{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				FieldName:    "name",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesDeleteEncryptedIndex_NoIdentity_NotAuthorizedError(t *testing.T) {
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
			// Starting all nodes with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users {
						name: String
					}
				`,
			},

			testUtils.DeleteEncryptedIndex{
				Identity:      testUtils.NoIdentity(),
				CollectionID:  0,
				FieldName:     "name",
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeDeleteEncryptedIndexPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesDeleteEncryptedIndex_NoIdentity_CLIandCandHTTPClient_NotAuthorizedError(t *testing.T) {
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
			// Starting all nodes with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users {
						name: String
					}
				`,
			},

			testUtils.DeleteEncryptedIndex{
				Identity:      testUtils.NoIdentity(),
				CollectionID:  0,
				FieldName:     "name",
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeGetCollectionPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesDeleteEncryptedIndex_WrongIdentity_NotAuthorizedError(t *testing.T) {
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
			// Starting all nodes with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users {
						name: String
					}
				`,
			},

			testUtils.DeleteEncryptedIndex{
				Identity:      testUtils.ClientIdentity(2),
				CollectionID:  0,
				FieldName:     "name",
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeDeleteEncryptedIndexPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesDeleteEncryptedIndex_WrongIdentity_CLIandCandHTTPClient_NotAuthorizedError(t *testing.T) {
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
			// Starting all nodes with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users {
						name: String
					}
				`,
			},

			testUtils.DeleteEncryptedIndex{
				Identity:      testUtils.ClientIdentity(2),
				CollectionID:  0,
				FieldName:     "name",
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeGetCollectionPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
