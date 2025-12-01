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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_Disabled_AdminRelation_DoesNotOwnTheDocument_CanNotAccessAndCanNotDACBypass(t *testing.T) {
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
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   examplePolicy,
			},
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema:   `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
			},
			testUtils.CreateDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},
			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			// This user, can not access the document when NAC is turned off, as DAC access gates it.
			// Note: no error here because DAC access is being used to gate here, not NAC.
			testUtils.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},

			// Grant dac-bypass access to user.
			// Note: since we are testing the temporarily disabled NAC case, we need to re-enable and
			// disable NAC otherwise we can't do the AddNACActorRelationship action.
			testUtils.ReEnableNAC{
				Identity: testUtils.ClientIdentity(1),
			},
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2), // Grant this user "admin" relation
				Relation:          "admin",
				ExpectedExistence: false,
			},
			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			// Note: This is a very important edge case to test. This user, still should not be able to bypass
			// or access the document even though they have dac-bypass permission (because NAC is disabled).
			// So, they have to go through DAC, hence can't access the document they don't own.
			testUtils.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_Disabled_AdminRelation_OwnThePrivateDocument_CanAccessButNotDACBypass(t *testing.T) {
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
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   examplePolicy,
			},
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema:   `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
			},
			// Temporarily disable to allow a non-node-owner to own some documents, and keep NAC disbaled
			// to test the disabled NAC case.
			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},
			testUtils.CreateDoc{
				Identity:     testUtils.ClientIdentity(2),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},

			// This user, can access the document as it has DAC ownership and NAC is disabled.
			testUtils.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},

			// Grant dac-bypass access to user.
			// Note: since we are testing the temporarily disabled NAC case, we need to re-enable and
			// disable NAC otherwise we can't do the AddNACActorRelationship action.
			testUtils.ReEnableNAC{
				Identity: testUtils.ClientIdentity(1),
			},
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2), // Grant this user "admin" relation
				Relation:          "admin",
				ExpectedExistence: false,
			},
			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			// This user, still should not be able to dac-bypass the document (because NAC is disabled).
			// However, since they do own the document and go through DAC, they continue to have access.
			testUtils.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_Disabled_AdminRelation_PublicDocument_CanAccessButNotDACBypass(t *testing.T) {
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
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   examplePolicy,
			},
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema:   `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
			},
			// Temporarily disable to make public document(s), and keep NAC disbaled to test the disabled NAC case.
			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},
			testUtils.CreateDoc{
				Identity:     testUtils.NoIdentity(),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},

			// This user, can access the document as it is a public document and NAC is disabled.
			testUtils.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},

			// Grant dac-bypass access to user.
			// Note: since we are testing the temporarily disabled NAC case, we need to re-enable and
			// disable NAC otherwise we can't do the AddNACActorRelationship action.
			testUtils.ReEnableNAC{
				Identity: testUtils.ClientIdentity(1),
			},
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2), // Grant this user "admin" relation
				Relation:          "admin",
				ExpectedExistence: false,
			},
			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},

			// This user, still should not be able to dac-bypass (because NAC is disabled).
			// However, since this is a puclic document not gated be DAC, they continue to have access.
			testUtils.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
