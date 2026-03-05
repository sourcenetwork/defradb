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
			&action.AddCollection{
				SDL: `
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
			&action.AddCollection{
				SDL: `
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
