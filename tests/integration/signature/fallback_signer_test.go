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

	"github.com/sourcenetwork/defradb/internal/merkle/clock"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/immutable"
)

func TestSignature_IfIdentityHasNoPrivateKeyButFallbackSignerIsSet_ShouldUseFallbackSigner(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning:  true,
		FallbackSigner: testUtils.NodeIdentity(0),
		// Fallback signer can be only tested with HTTP and CLI clients, because with Go client
		// when providing an identity, it includes the private key.
		SupportedClientTypes: immutable.Some([]testUtils.ClientType{
			testUtils.HTTPClientType,
			testUtils.CLIClientType,
		}),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						age: Int 
					}`,
			},
			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(0),
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
				Identity: testUtils.ClientIdentity(0),
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

func TestSignature_IfIdentityHasNoPrivateKey_ShouldFail(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		// Fallback signer can be only tested with HTTP and CLI clients, because with Go client
		// when providing an identity, it includes the private key.
		SupportedClientTypes: immutable.Some([]testUtils.ClientType{
			testUtils.HTTPClientType,
			testUtils.CLIClientType,
		}),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						age: Int 
					}`,
			},
			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(0),
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
				ExpectedError: clock.ErrIdentityWithoutPrivateKeyForSigning.Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
