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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestNAC_AdminRelation_CanVerifySignature(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				// Default signer can be only tested with HTTP and CLI clients, because with Go client
				// when providing an identity, it includes the private key.
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
					type Users {
						name: String
						age: Int 
					}`,
			},
			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(1),
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},

			// This user, can not perform this gated operation yet.
			testUtils.VerifyBlockSignature{
				Identity:       testUtils.ClientIdentity(2),
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreihymej6gbxq7qauy4tgt37di25uap2ahzq7z5d3ln3og5syo7rwmi",
				ExpectedError:  "not authorized to perform operation",
			},

			// Grant access to user.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2), // Grant this user "admin" relation
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// This user, can now perform this gated operation.
			testUtils.VerifyBlockSignature{
				Identity:       testUtils.ClientIdentity(2),
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreihymej6gbxq7qauy4tgt37di25uap2ahzq7z5d3ln3og5syo7rwmi",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_AdminRelation_GoClient_CanVerifySignature(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				// Creating of signed documents over HTTP is not supported yet, because signing
				// requires a private key which we do not pass over HTTP.
				state.GoClientType,
				state.CClientType,
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
					type Users {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(1),
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},

			// This user, can not perform this gated operation yet.
			testUtils.VerifyBlockSignature{
				Identity:       testUtils.ClientIdentity(2),
				SignerIdentity: testUtils.ClientIdentity(1).Value(),
				Cid:            "bafyreihymej6gbxq7qauy4tgt37di25uap2ahzq7z5d3ln3og5syo7rwmi",
				ExpectedError:  "not authorized to perform operation",
			},

			// Grant access to user.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2), // Grant this user "admin" relation
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// This user, can now perform this gated operation.
			testUtils.VerifyBlockSignature{
				Identity:       testUtils.ClientIdentity(2),
				SignerIdentity: testUtils.ClientIdentity(1).Value(),
				Cid:            "bafyreihymej6gbxq7qauy4tgt37di25uap2ahzq7z5d3ln3og5syo7rwmi",
				ExpectedError:  "could not find",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
