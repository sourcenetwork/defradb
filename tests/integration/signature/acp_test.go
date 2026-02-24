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

	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

const policy = `
name: test
description: a test policy which marks a collection in a database as a resource

resources:
- name: users
  permissions:
  - name: read
    expr: reader
  - name: update
  - name: delete

  relations:
  - name: reader
    types:
    - actor

  - name: admin
    manages:
    - reader
    types:
    - actor
`

func TestSignatureACP_IfHasNoAccessToDoc_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		SupportedClientTypes: immutable.Some([]state.ClientType{
			// Creating of signed documents over HTTP is not supported yet, because signing
			// requires a private key which we do not pass over HTTP.
			state.GoClientType,
		}),
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   policy,
			},
			&action.AddSchema{
				Schema: `
 					type Users @policy(
						id: "{{.Policy0}}",
 						resource: "users"
 					) {
 						name: String
 						age: Int
 					}
 				`,
			},
			&action.AddDoc{
				Identity: testUtils.ClientIdentity(1),
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			testUtils.VerifyBlockSignature{
				Identity:       testUtils.NodeIdentity(1),
				SignerIdentity: testUtils.ClientIdentity(1).Value(),
				Cid:            "bafyreia5uzkhoqvhccljbpiiafrjyvperxphmun264ul6esvuosk6pnf5m",
				ExpectedError:  db.ErrMissingPermission.Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSignatureACP_IfHasAccessToDoc_ValidateSignature(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		SupportedClientTypes: immutable.Some([]state.ClientType{
			// Creating of signed documents over HTTP is not supported yet, because signing
			// requires a private key which we do not pass over HTTP.
			state.GoClientType,
		}),
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   policy,
			},
			&action.AddSchema{
				Schema: `
 					type Users @policy(
						id: "{{.Policy0}}",
 						resource: "users"
 					) {
 						name: String
 						age: Int
 					}
 				`,
			},
			&action.AddDoc{
				Identity: testUtils.ClientIdentity(1),
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			testUtils.VerifyBlockSignature{
				Identity:       testUtils.ClientIdentity(1),
				SignerIdentity: testUtils.ClientIdentity(1).Value(),
				Cid:            "bafyreia5uzkhoqvhccljbpiiafrjyvperxphmun264ul6esvuosk6pnf5m",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
