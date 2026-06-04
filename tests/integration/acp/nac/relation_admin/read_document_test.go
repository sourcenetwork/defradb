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

package test_acp_nac_relation_admin

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_AdminRelation_CanReadDocument(t *testing.T) {
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

			// This user, can not perform this gated operation yet.
			&action.Request{
				Identity:      testUtils.ClientIdentity(2),
				Request:       `query{ User { name } }`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeGetCollectionPerm),
			},

			// Grant access to user.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2), // Grant this user "admin" relation
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// This user, can now perform this gated operation.
			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request:  `query{ User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Shahzad"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_AdminRelation_MaterializedView_CanReadDocument(t *testing.T) {
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

			// Grant access to user.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2), // Grant this user "admin" relation
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// This user, can now perform this gated operation.
			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request:  `query{ User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Shahzad"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
