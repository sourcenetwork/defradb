// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_aac_setup_then_start

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestAAC_GatesDeletingDACRelationshipPreSetup_AllowIfAuthorizedElseError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "admin acp correctly gates deleting DAC relationship operation (setup before aac), allow if authorized, otherwise error",
		SupportedDatabaseTypes: immutable.Some(
			[]testUtils.DatabaseType{
				// This test only supports file type databases since the setup steps will be done before
				// the node is re-started with aac enabled (if it's in-memory it will loose setup state).
				testUtils.BadgerFileType,
			},
		),
		Actions: []any{
			// Note: Since this is not an in-memory test, we can do the setup steps before aac is enabled.
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy: `
                    name: Test Policy
                    description: A Policy
                    actor:
                      name: actor
                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner + reader + updater + deleter
                          update:
                            expr: owner + updater
                          delete:
                            expr: owner + deleter
                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                          updater:
                            types:
                              - actor
                          deleter:
                            types:
                              - actor
                `,
			},
			testUtils.SchemaUpdate{
				Schema:  `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
				Replace: map[string]testUtils.ReplaceType{"Policy0": testUtils.NewPolicyIndex(0)},
			},
			testUtils.CreateDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(3),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedExistence: false,
			},

			// Starting with ACC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			// We haven't authorized non-identities. So, this should error.
			testUtils.DeleteDACActorRelationship{
				RequestorIdentity: testUtils.NoIdentity(),
				TargetIdentity:    testUtils.ClientIdentity(3),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedError:     "not authorized to perform operation",
			},

			// Wrong user/identity will also not be authorized.
			testUtils.DeleteDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2),
				TargetIdentity:    testUtils.ClientIdentity(3),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedError:     "not authorized to perform operation",
			},

			// This should work as the identity is authorized.
			testUtils.DeleteDACActorRelationship{
				RequestorIdentity:   testUtils.ClientIdentity(1),
				TargetIdentity:      testUtils.ClientIdentity(3),
				CollectionID:        0,
				DocID:               0,
				Relation:            "reader",
				ExpectedRecordFound: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
