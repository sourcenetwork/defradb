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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_AdminRelation_CanCollectionGetByName(t *testing.T) {
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
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type Users {}
				`,
			},

			// This user, can not perform this gated operation yet.
			&action.GetCollections{
				Identity: testUtils.ClientIdentity(2),
				FilterOptions: client.CollectionFetchOptions{
					Name:            immutable.Some("Users"),
					IncludeInactive: immutable.Some(false),
				},
				ExpectedError: "not authorized to perform operation",
			},

			// Grant access to user.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2), // Grant this user "admin" relation
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// This user, can now perform this gated operation.
			&action.GetCollections{
				Identity: testUtils.ClientIdentity(2),
				FilterOptions: client.CollectionFetchOptions{
					Name:            immutable.Some("Users"),
					IncludeInactive: immutable.Some(false),
				},
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
