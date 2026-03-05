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

	"github.com/sourcenetwork/immutable"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_WithDACEnabled_AccessByEmptyUser_PrivateDocumentOwnedByNodeOwner_NotAuthorizedError(t *testing.T) {
	// todo: Investigate and test this behavior across all client types when implementing granular NAC permissions.
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
			// Make a private document.
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   examplePolicy,
			},
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL:      `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},

			// Empty user can not access the private document.
			&action.Request{
				Identity: testUtils.NoIdentity(),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeGetCollectionPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_WithDACEnabled_AccessByEmptyUser_PrivateDocumentOwnedByNodeOwner_MaterializedView_NotAuthorizedError(t *testing.T) {
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
			// Make a private document.
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   examplePolicy,
			},
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL:      `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},

			// With materialized views, the view refresh gate is hit first.
			&action.Request{
				Identity: testUtils.NoIdentity(),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeRefreshViewPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_WithDACEnabled_AccessByEmptyUser_PrivateDocumentOwnedByNonNodeOwner_NotAuthorizedError(t *testing.T) {
	// todo: Investigate and test this behavior across all client types when implementing granular NAC permissions.
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
			// Temporarily disable to allow a non-node-owner to own some documents.
			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},
			// Make a private document.
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(2),
				Policy:   examplePolicy,
			},
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(2),
				SDL:      `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(2),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},
			testUtils.ReEnableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			// Empty user can not access the private document.
			&action.Request{
				Identity: testUtils.NoIdentity(),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeGetCollectionPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_WithDACEnabled_AccessByEmptyUser_PrivateDocumentOwnedByNonNodeOwner_MaterializedView_NotAuthorizedError(t *testing.T) {
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
			// Temporarily disable to allow a non-node-owner to own some documents.
			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},
			// Make a private document.
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(2),
				Policy:   examplePolicy,
			},
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(2),
				SDL:      `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(2),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},
			testUtils.ReEnableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			// With materialized views, the view refresh gate is hit first.
			&action.Request{
				Identity: testUtils.NoIdentity(),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeRefreshViewPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_WithDACEnabled_AccessEmptyUser_PublicDocument_NotAuthorizedError(t *testing.T) {
	// todo: Investigate and test this behavior across all client types when implementing granular NAC permissions.
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
			// Temporarily disable to allow easy creation of public document.
			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},
			// Make a non-node-owner own a document.
			testUtils.AddDACPolicy{
				Identity: testUtils.NodeIdentity(0), // Doesn't matter who adds the policy.
				Policy:   examplePolicy,
			},
			&action.AddCollection{
				Identity: testUtils.NoIdentity(),
				SDL:      `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.NoIdentity(),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},
			testUtils.ReEnableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			// Empty user can not access the private document, because of NAC.
			&action.Request{
				Identity: testUtils.NoIdentity(),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeGetCollectionPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_WithDACEnabled_AccessEmptyUser_PublicDocument_MaterializedView_NotAuthorizedError(t *testing.T) {
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
			// Temporarily disable to allow easy creation of public document.
			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},
			// Make a non-node-owner own a document.
			testUtils.AddDACPolicy{
				Identity: testUtils.NodeIdentity(0), // Doesn't matter who adds the policy.
				Policy:   examplePolicy,
			},
			&action.AddCollection{
				Identity: testUtils.NoIdentity(),
				SDL:      `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.NoIdentity(),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},
			testUtils.ReEnableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			// With materialized views, the view refresh gate is hit first.
			&action.Request{
				Identity: testUtils.NoIdentity(),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeRefreshViewPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
