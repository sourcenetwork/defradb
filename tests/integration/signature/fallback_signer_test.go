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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestSignature_IfIdentityHasNoPrivateKey_ShouldUseNodeIdentity(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		// Default signer can be only tested with HTTP and CLI clients, because with Go client
		// when providing an identity, it includes the private key.
		SupportedClientTypes: immutable.Some([]state.ClientType{
			state.HTTPClientType,
			state.CLIClientType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int 
					}`,
			},
			&action.AddDoc{
				Identity: testUtils.ClientIdentity(0),
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreib7pek5q7pweyljoylz7qusofwls3wnnzik6j5ibo3pe77fz4iw3e",
			},
			testUtils.UpdateDoc{
				Identity: testUtils.ClientIdentity(0),
				Doc: `{
					"age": 23
				}`,
			},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreiehh36sndwjshkpatwir3cqhfi7vya22kdsst65oyp5ejufbf7m4e",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
