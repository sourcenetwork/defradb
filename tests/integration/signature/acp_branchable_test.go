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
	"github.com/sourcenetwork/defradb/tests/multiplier"
	"github.com/sourcenetwork/defradb/tests/state"
)

// branchableACPSDL is a branchable, permissioned collection. Registering it (with an identity)
// makes its collection-level commit DAG private, owned by the registering identity.
const branchableACPSDL = `
	type Users @branchable @policy(
		id: "{{.Policy0}}",
		resource: "users"
	) {
		name: String
		age: Int
	}
`

// The collection-level commit of a branchable collection carries no docID. Verifying its signature
// must be gated on the collection object (the path added for branchable collections in
// [db.DB.VerifySignature]); an actor without read access to the collection is refused.
func TestSignatureACP_WithBranchableCollection_IfHasNoCollectionAccess_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		// hardcoded block CID would change under encryption
		MultiplierExcludes: []string{multiplier.EncryptedDocs},
		EnableSigning:      true,
		SupportedClientTypes: immutable.Some([]state.ClientType{
			// Creating of signed documents over HTTP is not supported yet, because signing
			// requires a private key which we do not pass over HTTP.
			state.GoClientType,
			state.CClientType,
		}),
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   policy,
			},
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL:      branchableACPSDL,
			},
			&action.AddDoc{
				Identity: testUtils.ClientIdentity(1),
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			// The collection-level commit (fieldName: nil, no docID).
			testUtils.VerifyBlockSignature{
				Identity:       testUtils.ClientIdentity(2),
				SignerIdentity: testUtils.ClientIdentity(1).Value(),
				Cid:            "bafyreia7dgbk4gjeo3hvdqcbizzl4f25onmy7oetwhvxe6sq3lup5n7xju",
				ExpectedError:  db.ErrMissingPermission.Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// The collection owner has read access to the collection object, so it may verify the signature of
// the collection-level commit (and, additively, the document commits that make up the DAG).
func TestSignatureACP_WithBranchableCollection_IfHasCollectionAccess_ValidateSignature(t *testing.T) {
	test := testUtils.TestCase{
		// hardcoded block CID would change under encryption
		MultiplierExcludes: []string{multiplier.EncryptedDocs},
		EnableSigning:      true,
		SupportedClientTypes: immutable.Some([]state.ClientType{
			// Creating of signed documents over HTTP is not supported yet, because signing
			// requires a private key which we do not pass over HTTP.
			state.GoClientType,
			state.CClientType,
		}),
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   policy,
			},
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL:      branchableACPSDL,
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
				Cid:            "bafyreia7dgbk4gjeo3hvdqcbizzl4f25onmy7oetwhvxe6sq3lup5n7xju",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
