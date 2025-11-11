// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package signature

import (
	"testing"

	"github.com/sourcenetwork/defradb/crypto"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestSignatureVerify_WithValidData_ShouldVerify(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						age: Int 
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreibphszakimmug77fvftqmpv4uqtn3rmc5rv4u6qafeqiuu7oyeyca",
			},
			testUtils.UpdateDoc{
				Doc: `{
					"age": 23
				}`,
			},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreiemskbirmphqngy3kivou4e4phcuiffcdql4mvqp3twmiipwt3pta",
			},
			testUtils.DeleteDoc{},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreifv2o6ywzdxkbvfxlwtks4u3nfplprnbakl663rykwp3zscx5b3ay",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSignatureVerify_WithDifferentKeyType_ShouldVerify(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		IdentityTypes: map[state.Identity]crypto.KeyType{
			testUtils.NodeIdentity(0).Value(): crypto.KeyTypeEd25519,
		},
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						age: Int 
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreihvuvkk6qa2rmnujlwwpbvzd6nxkahqmlv22n4x7hb6q4obmf7euy",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSignatureVerify_WithWrongIdentity_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						age: Int 
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(1).Value(),
				Cid:            "bafyreibphszakimmug77fvftqmpv4uqtn3rmc5rv4u6qafeqiuu7oyeyca",
				ExpectedError:  coreblock.ErrSignaturePubKeyMismatch.Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSignatureVerify_WithWrongCid_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						age: Int 
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreiemskbirmphqngy3kivou4e4phcuiffcdql4mvqp3twmiipwt3pta",
				ExpectedError:  "could not find",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
