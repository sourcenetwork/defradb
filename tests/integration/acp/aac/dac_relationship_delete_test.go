// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_aac

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestAAC_GatesDeletingDACRelationship_AllowIfAuthorizedElseError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "admin acp correctly gates deleting DAC relationship operation, allow if authorized, otherwise error",
		Actions: []any{
			// Starting with ACC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},
			// Note: Doing setup steps after starting with aac enabled, otherwise the in-memory tests
			// will loose setup state when the restart happens (i.e. the restart that started aac).
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
				Identity: testUtils.ClientIdentity(1),
				Schema:   `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
				Replace:  map[string]testUtils.ReplaceType{"Policy0": testUtils.NewPolicyIndex(0)},
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
			// Note: Setup is now done, the test code that follows is what we want to assert.

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
