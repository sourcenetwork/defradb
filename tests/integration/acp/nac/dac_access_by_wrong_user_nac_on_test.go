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

func TestNAC_WithDACEnabled_AccessByNonNodeOwner_OwnsTheDocument_NotAuthorizedError(t *testing.T) {
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
			// Temporarily disable to allow a non-node-owner to own some documents.
			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},
			// Make a non-node-owner own a document.
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

			// Document owner but can not access, because NAC takes precedence.
			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeCollectionGetPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_WithDACEnabled_AccessByNonNodeOwner_OwnsTheDocument_MaterializedView_NotAuthorizedError(t *testing.T) {
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
			// Make a non-node-owner own a document.
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
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeViewRefreshPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_WithDACEnabled_AccessByNonNodeOwner_DoesNotOwnTheDocument_NotAuthorizedError(t *testing.T) {
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
			// Temporarily disable to allow a non-node-owner to own some documents.
			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},
			// Make a non-node-owner own a document.
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

			// Not authorized user, can not access because NAC is enabled.
			&action.Request{
				Identity: testUtils.ClientIdentity(3),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeCollectionGetPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_WithDACEnabled_AccessByNonNodeOwner_DoesNotOwnTheDocument_MaterializedView_NotAuthorizedError(t *testing.T) {
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
			// Make a non-node-owner own a document.
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
				Identity: testUtils.ClientIdentity(3),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeViewRefreshPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_WithDACEnabled_AccessByNonNodeOwner_PublicDocument_AllowAccess(t *testing.T) {
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

			// Not authorized user, can not access even public document, because NAC is enabled.
			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeCollectionGetPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_WithDACEnabled_AccessByNonNodeOwner_PublicDocument_MaterializedView_NotAuthorizedError(t *testing.T) {
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
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeViewRefreshPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
