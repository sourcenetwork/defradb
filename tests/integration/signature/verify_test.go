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
				Cid:            "{{.CID0_0_0}}",
			},
			&action.UpdateDoc{
				Doc: `{
					"age": 23
				}`,
			},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "{{.CID0_0_1}}",
			},
			testUtils.DeleteDoc{},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "{{.CID0_0_2}}",
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
				Cid:            "{{.CID0_0_0}}",
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
				Cid:            "{{.CID0_0_0}}",
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
