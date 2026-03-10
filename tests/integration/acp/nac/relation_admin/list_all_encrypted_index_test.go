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

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_AdminRelation_CanListAllEncryptedIndex(t *testing.T) {
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
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeListAllEncryptedIndexPerm),
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
