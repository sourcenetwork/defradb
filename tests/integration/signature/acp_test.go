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

	"github.com/sourcenetwork/defradb/internal/db"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

const policy = `
	name: test
	description: a test policy which marks a collection in a database as a resource

	actor:
	  name: actor

	resources:
	  users:
		permissions:
		  read:
			expr: owner + reader
		  update:
			expr: owner
		  delete:
			expr: owner

		relations:
		  owner:
			types:
			  - actor
		  reader:
			types:
			  - actor
		  admin:
			manages:
			  - reader
			types:
			  - actor
`

func TestSignatureACP_IfHasNoAccessToDoc_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		Actions: []any{
			testUtils.AddPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   policy,
			},
			testUtils.SchemaUpdate{
				Schema: `
 					type Users @policy(
						id: "{{.Policy0}}",
 						resource: "users"
 					) {
 						name: String
 						age: Int
 					}
 				`,
				Replace: map[string]testUtils.ReplaceType{
					"Policy0": testUtils.NewPolicyIndex(0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			testUtils.VerifyBlockSignature{
				Identity:       testUtils.NodeIdentity(1),
				SignerIdentity: testUtils.ClientIdentity(1).Value(),
				Cid:            "bafyreiaqqaqoe73ioolf6lofprgekb4lnrcteanpbjgjegkn6ug77ghmri",
				ExpectedError:  db.ErrMissingPermission.Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSignatureACP_IfHasAccessToDoc_ValidateSignature(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		Actions: []any{
			testUtils.AddPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   policy,
			},
			testUtils.SchemaUpdate{
				Schema: `
 					type Users @policy(
						id: "{{.Policy0}}",
 						resource: "users"
 					) {
 						name: String
 						age: Int
 					}
 				`,
				Replace: map[string]testUtils.ReplaceType{
					"Policy0": testUtils.NewPolicyIndex(0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			testUtils.VerifyBlockSignature{
				Identity:       testUtils.ClientIdentity(1),
				SignerIdentity: testUtils.ClientIdentity(1).Value(),
				Cid:            "bafyreiaqqaqoe73ioolf6lofprgekb4lnrcteanpbjgjegkn6ug77ghmri",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
