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
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_AdminRelation_CanPurgeDocument(t *testing.T) {
	test := testUtils.TestCase{Actions: []any{
		testUtils.Close{},
		testUtils.Start{Identity: testUtils.ClientIdentity(1), EnableNAC: true},
		&action.AddCollection{
			Identity: testUtils.ClientIdentity(1),
			SDL:      `type Users { name: String }`,
		},
		&action.AddDoc{
			Identity:     testUtils.ClientIdentity(1),
			CollectionID: 0,
			Doc:          `{"name":"alice"}`,
		},
		&action.PurgeDocs{
			Identity:           testUtils.ClientIdentity(2),
			CollectionIdentity: testUtils.ClientIdentity(1),
			CollectionIndex:    0,
			DocIndexes:         []int{0},
			ExpectedError:      testUtils.FormatExpectedErrorWithPermission(acpTypes.NodePurgeDocumentPerm),
		},
		testUtils.AddNACActorRelationship{
			RequestorIdentity: testUtils.ClientIdentity(1),
			TargetIdentity:    testUtils.ClientIdentity(2),
			Relation:          "admin",
			ExpectedExistence: false,
		},
		&action.PurgeDocs{
			Identity:        testUtils.ClientIdentity(2),
			CollectionIndex: 0,
			DocIndexes:      []int{0},
		},
	}}

	testUtils.ExecuteTestCase(t, test)
}
