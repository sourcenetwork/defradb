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
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSignatureVerify_WithValidData_ShouldVerify(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		Actions: []any{
			testUtils.SchemaUpdate{
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
				Cid:            "bafyreicwhd5s762awsrx6eowwqkkfpq7r5nnjosiru7blgaxo32wx6enp4",
			},
			testUtils.UpdateDoc{
				Doc: `{
					"age": 23
				}`,
			},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreidenvkbjuqismfbng463tfxsjmapvnvdyh4hmdx74ec5skj63ma2a",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSignatureVerify_WithDifferentKeyType_ShouldVerify(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		IdentityTypes: map[testUtils.Identity]crypto.KeyType{
			testUtils.NodeIdentity(0).Value(): crypto.KeyTypeEd25519,
		},
		Actions: []any{
			testUtils.SchemaUpdate{
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
				Cid:            "bafyreifawqsvcshb77gorqbnopnyg6mi6ft3iz7begvlhm67hnkhyqnfla",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSignatureVerify_WithWrongIdentity_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		Actions: []any{
			testUtils.SchemaUpdate{
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
				Cid:            "bafyreicwhd5s762awsrx6eowwqkkfpq7r5nnjosiru7blgaxo32wx6enp4",
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
			testUtils.SchemaUpdate{
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
				Cid:            "bafyreidenvkbjuqismfbng463tfxsjmapvnvdyh4hmdx74ec5skj63ma2a",
				ExpectedError:  "could not find",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
