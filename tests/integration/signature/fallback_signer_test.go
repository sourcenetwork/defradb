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
			testUtils.HTTPClientType,
			testUtils.CLIClientType,
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
				Identity: testUtils.ClientIdentity(0),
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
				Identity: testUtils.ClientIdentity(0),
				Doc: `{
					"age": 23
				}`,
			},
			testUtils.VerifyBlockSignature{
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreidazqqhrmdd6wnx33obcnd67vce33qbbmdbmzkofpfu77qach36sq",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
