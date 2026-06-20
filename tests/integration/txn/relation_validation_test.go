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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

// SDL definition order determines CollectionID:
//
//	CollectionID 0 = Company (first type in SDL)
//	CollectionID 1 = Employee (second type in SDL)
const companyEmployeeSDL = `
	type Company {
		name: String
	}
	type Employee {
		name: String
		company: Company
	}
`

// TestTxnRelation_CreateTargetAndLinkInSameTxn_NoError asserts that a document
// and its relation target can both be created within a single transaction: the
// relation validator can read the target from the same transaction's buffer.
func TestTxnRelation_CreateTargetAndLinkInSameTxn_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: companyEmployeeSDL,
			},
			// Create Company in Txn1 (not yet committed).
			&action.AddDoc{
				CollectionID:  0,
				TransactionID: immutable.Some(1),
				Doc:           `{"name": "Acme"}`,
			},
			// Create Employee in the same Txn1 — validateRelationDocIDs reads from
			// Txn1's buffer and sees the Company even though it isn't committed yet.
			&action.AddDoc{
				CollectionID:  1,
				TransactionID: immutable.Some(1),
				DocMap: map[string]any{
					"name":    "Alice",
					"company": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CommitTransaction{
				TransactionID: 1,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestTxnRelation_LinkInTxnBeforeTargetCommitted_Error asserts that a transaction
// cannot link to a document that was created in a different, uncommitted transaction:
// the relation validator does not see the other transaction's uncommitted data.
func TestTxnRelation_LinkInTxnBeforeTargetCommitted_Error(t *testing.T) {
	test := testUtils.TestCase{
		// LevelDB does not support concurrent transactions.
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: companyEmployeeSDL,
			},
			// Create Company in Txn1 (NOT committed).
			&action.AddDoc{
				CollectionID:  0,
				TransactionID: immutable.Some(1),
				Doc:           `{"name": "Acme"}`,
			},
			// Create Employee in Txn2 — Txn2 cannot see Txn1's uncommitted Company.
			&action.AddDoc{
				CollectionID:  1,
				TransactionID: immutable.Some(2),
				DocMap: map[string]any{
					"name":    "Alice",
					"company": testUtils.NewDocIndex(0, 0),
				},
				ExpectedError: "relation target document not found",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestTxnRelation_DeleteTargetThenLink_Error asserts that when a relation target is
// soft-deleted within a transaction, a subsequent AddDoc in the same transaction that
// references the deleted target is rejected by the relation validator.
func TestTxnRelation_DeleteTargetThenLink_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: companyEmployeeSDL,
			},
			// Create and commit Company outside any transaction.
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Acme"}`,
			},
			// In Txn1: soft-delete the Company.
			testUtils.DeleteDoc{
				CollectionID:  0,
				DocID:         0,
				TransactionID: immutable.Some(1),
			},
			// In Txn1: attempt to create Employee linking to the now-deleted Company.
			&action.AddDoc{
				CollectionID:  1,
				TransactionID: immutable.Some(1),
				DocMap: map[string]any{
					"name":    "Alice",
					"company": testUtils.NewDocIndex(0, 0),
				},
				ExpectedError: "relation target document not found",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
