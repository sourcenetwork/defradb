// Copyright 2026 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_AdminRelation_CanEncryptedIndexListAll(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			// This user, can not perform this gated operation yet.
			testUtils.ListAllEncryptedIndexes{
				Identity:      testUtils.ClientIdentity(2),
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeEncryptedIndexListAllPerm),
			},

			// Grant access to user.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// This user, can now perform this gated operation.
			testUtils.ListAllEncryptedIndexes{
				Identity:        testUtils.ClientIdentity(2),
				ExpectedIndexes: map[client.CollectionName][]client.EncryptedIndexDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
