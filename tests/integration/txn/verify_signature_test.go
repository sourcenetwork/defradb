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

package txn_testing

import (
	"testing"

	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

// This test runs VerifyBlockSignature inside of a transaction, illustrating that it works.
func TestTxn_VerifyBlockSignature_InsideTxn_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		// LevelDB does not support concurrent transactions
		// todo: https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		EnableSigning: true,
		IdentityTypes: map[state.Identity]crypto.KeyType{
			testUtils.NodeIdentity(0).Value(): crypto.KeyTypeEd25519,
		},
		Actions: []any{
			&action.AddCollection{
				TransactionID: immutable.Some(1),
				SDL: `
					type Users {
						name: String
						age: Int 
					}`,
			},
			&action.AddDoc{
				TransactionID: immutable.Some(1),
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			testUtils.VerifyBlockSignature{
				TransactionID:  immutable.Some(1),
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreibxlg2hmbbhbia4zywlif4xhozrf47js6r46ag5bcw72uc5m53csi",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs VerifyBlockSignature outside of a transaction containing the block it wants to
// verify, illustrating transactional isolation.
func TestTxn_VerifyBlockSignature_OutsideTxn_Fails(t *testing.T) {
	test := testUtils.TestCase{
		// LevelDB does not support concurrent transactions
		// todo: https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		EnableSigning: true,
		IdentityTypes: map[state.Identity]crypto.KeyType{
			testUtils.NodeIdentity(0).Value(): crypto.KeyTypeEd25519,
		},
		Actions: []any{
			&action.AddCollection{
				TransactionID: immutable.Some(1),
				SDL: `
					type Users {
						name: String
						age: Int 
					}`,
			},
			&action.AddDoc{
				TransactionID: immutable.Some(1),
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			testUtils.VerifyBlockSignature{
				TransactionID:  immutable.Some(2),
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreibxlg2hmbbhbia4zywlif4xhozrf47js6r46ag5bcw72uc5m53csi",
				ExpectedError:  "ipld: could not find bafyreibxlg2hmbbhbia4zywlif4xhozrf47js6r46ag5bcw72uc5m53csi",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
