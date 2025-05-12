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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/crypto"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestSignatureVerify_WithValidData_ShouldVerify(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		SupportedClientTypes: immutable.Some([]state.ClientType{
			// TODO: C binding test harness must be reworked to support this test
			// See: https://github.com/sourcenetwork/defradb/issues/3919
			testUtils.GoClientType,
			testUtils.CLIClientType,
			testUtils.HTTPClientType,
			testUtils.JSClientType,
		}),
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
				Cid:            "bafyreihxhybgbd5vjpoobyrol4unb5p5jy4jy445sf4hcvq5rbo2h56hce",
			},
			testUtils.UpdateDoc{
				Doc: `{
					"age": 23
				}`,
			},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreidazqqhrmdd6wnx33obcnd67vce33qbbmdbmzkofpfu77qach36sq",
			},
			testUtils.DeleteDoc{},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreifr7ukfh7lvbbmv2uh6lvd2zvr3x4kuqq73a35awya4smbyeyprum",
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
		SupportedClientTypes: immutable.Some([]state.ClientType{
			// TODO: C binding test harness must be reworked to support this test
			// See: https://github.com/sourcenetwork/defradb/issues/3919
			testUtils.GoClientType,
			testUtils.CLIClientType,
			testUtils.HTTPClientType,
			testUtils.JSClientType,
		}),
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
				Cid:            "bafyreiabwxmlv6fbqb2acmqiifolz52ztxyctxkynvm7cleghf5nqqexcq",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSignatureVerify_WithWrongIdentity_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		SupportedClientTypes: immutable.Some([]state.ClientType{
			// TODO: C binding test harness must be reworked to support this test
			// See: https://github.com/sourcenetwork/defradb/issues/3919
			testUtils.GoClientType,
			testUtils.CLIClientType,
			testUtils.HTTPClientType,
			testUtils.JSClientType,
		}),
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
				Cid:            "bafyreihxhybgbd5vjpoobyrol4unb5p5jy4jy445sf4hcvq5rbo2h56hce",
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
				Cid:            "bafyreidazqqhrmdd6wnx33obcnd67vce33qbbmdbmzkofpfu77qach36sq",
				ExpectedError:  "could not find",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
