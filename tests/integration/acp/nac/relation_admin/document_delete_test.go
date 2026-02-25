// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_nac_relation_admin

import (
	"testing"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

func TestNAC_AdminRelation_CanDocumentDelete(t *testing.T) {
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
						name: String
						age: Int 
					}`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc: `{
					"name": "Shahzad",
					"age": 48
				}`,
			},

			// This user, can not perform this gated operation yet.
			// This should work as the identity is authorized.
			testUtils.DeleteDoc{
				Identity:      testUtils.ClientIdentity(2),
				CollectionID:  0,
				DocID:         0,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeDocumentDeletePerm),
			},

			// Grant access to user.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2), // Grant this user "admin" relation
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// This user, can now perform this gated operation.
			testUtils.DeleteDoc{
				Identity:     testUtils.ClientIdentity(2),
				CollectionID: 0,
				DocID:        0,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_AdminRelation_CLIandCandHTTPClient_CanDocumentDelete(t *testing.T) {
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
						name: String
						age: Int 
					}`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc: `{
					"name": "Shahzad",
					"age": 48
				}`,
			},

			// This user, can not perform this gated operation yet.
			// This should work as the identity is authorized.
			testUtils.DeleteDoc{
				Identity:      testUtils.ClientIdentity(2),
				CollectionID:  0,
				DocID:         0,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeCollectionGetPerm),
			},

			// Grant access to user.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2), // Grant this user "admin" relation
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// This user, can now perform this gated operation.
			testUtils.DeleteDoc{
				Identity:     testUtils.ClientIdentity(2),
				CollectionID: 0,
				DocID:        0,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
