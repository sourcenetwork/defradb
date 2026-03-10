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
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestNAC_AdminRelation_CanDeleteEncryptedIndex(t *testing.T) {
	test := testUtils.TestCase{
		// todo: Investigate and test this behavior across all client types when implementing granular NAC permissions.
		// See: https://github.com/sourcenetwork/defradb/issues/4383
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.GoClientType,
			},
		),
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
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

			// This user, can not perform this gated operation yet.
			testUtils.DeleteEncryptedIndex{
				Identity:      testUtils.ClientIdentity(2),
				CollectionID:  0,
				FieldName:     "name",
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeDeleteEncryptedIndexPerm),
			},

			// Grant access to user.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedExistence: false,
			},

			testUtils.NewEncryptedIndex{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				FieldName:    "name",
			},

			// This user, can now perform this gated operation.
			testUtils.DeleteEncryptedIndex{
				Identity:     testUtils.ClientIdentity(2),
				CollectionID: 0,
				FieldName:    "name",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_AdminRelation_CLIandCandHTTPClient_CanDeleteEncryptedIndex(t *testing.T) {
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
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users {
						name: String
					}
				`,
			},

			// This user, can not perform this gated operation yet.
			testUtils.DeleteEncryptedIndex{
				Identity:      testUtils.ClientIdentity(2),
				CollectionID:  0,
				FieldName:     "name",
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeGetCollectionPerm),
			},

			// Grant access to user.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedExistence: false,
			},

			testUtils.NewEncryptedIndex{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				FieldName:    "name",
			},

			// This user, can now perform this gated operation.
			testUtils.DeleteEncryptedIndex{
				Identity:     testUtils.ClientIdentity(2),
				CollectionID: 0,
				FieldName:    "name",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
