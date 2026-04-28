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
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int 
					}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreihymej6gbxq7qauy4tgt37di25uap2ahzq7z5d3ln3og5syo7rwmi",
			},
			&action.UpdateDoc{
				Doc: `{
					"age": 23
				}`,
			},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreichuvsbsr3oo4xeqfi55mrh4us77z2bg2foemuzhn5idomya6epl4",
			},
			testUtils.DeleteDoc{},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreigq4hkl7kgcj6qssol4ms3spagjjlaume2xatogdxqxc3h45td6q4",
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
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int 
					}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreibxlg2hmbbhbia4zywlif4xhozrf47js6r46ag5bcw72uc5m53csi",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSignatureVerify_WithWrongIdentity_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int 
					}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(1).Value(),
				Cid:            "bafyreihymej6gbxq7qauy4tgt37di25uap2ahzq7z5d3ln3og5syo7rwmi",
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
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int 
					}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreichuvsbsr3oo4xeqfi55mrh4us77z2bg2foemuzhn5idomya6epl4",
				ExpectedError:  "could not find",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
