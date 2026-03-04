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
)

func TestNAC_GatesReadDocument_AuthorizedIdentity_AllowAccess(t *testing.T) {
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
				SDL:      `type User { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},

			// This should work as the identity is authorized.
			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request:  `query{ User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Shahzad"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesReadDocument_NoIdentity_NotAuthorizedError(t *testing.T) {
	// todo: Investigate and test this behavior across all view types when implementing granular NAC permissions.
	// See: https://github.com/sourcenetwork/defradb/issues/4383
	test := testUtils.TestCase{
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{testUtils.CachelessViewType}),
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
				SDL:      `type User { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},

			// This should work as the identity is authorized.
			&action.Request{
				Identity:      testUtils.NoIdentity(),
				Request:       `query{ User { name } }`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeGetCollectionPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesReadDocument_NoIdentity_MaterializedView_NotAuthorizedError(t *testing.T) {
	// todo: Investigate and test this behavior across all view types when implementing granular NAC permissions.
	// See: https://github.com/sourcenetwork/defradb/issues/4383
	test := testUtils.TestCase{
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{testUtils.MaterializedViewType}),
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
				SDL:      `type User { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},

			// With materialized views, the view refresh gate is hit first.
			&action.Request{
				Identity:      testUtils.NoIdentity(),
				Request:       `query{ User { name } }`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeRefreshViewPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesReadDocument_WrongIdentity_NotAuthorizedError(t *testing.T) {
	// todo: Investigate and test this behavior across all view types when implementing granular NAC permissions.
	// See: https://github.com/sourcenetwork/defradb/issues/4383
	test := testUtils.TestCase{
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{testUtils.CachelessViewType}),
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
				SDL:      `type User { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},

			// This should work as the identity is authorized.
			&action.Request{
				Identity:      testUtils.ClientIdentity(2),
				Request:       `query{ User { name } }`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeGetCollectionPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesReadDocument_WrongIdentity_MaterializedView_NotAuthorizedError(t *testing.T) {
	// todo: Investigate and test this behavior across all view types when implementing granular NAC permissions.
	// See: https://github.com/sourcenetwork/defradb/issues/4383
	test := testUtils.TestCase{
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{testUtils.MaterializedViewType}),
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
				SDL:      `type User { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},

			// With materialized views, the view refresh gate is hit first.
			&action.Request{
				Identity:      testUtils.ClientIdentity(2),
				Request:       `query{ User { name } }`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeRefreshViewPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
