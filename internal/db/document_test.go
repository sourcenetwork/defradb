// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

// employeeCompanySchema defines a two-collection schema used across all
// document validation unit tests.
const employeeCompanySchema = `
	type Company {
		name: String
	}
	type Employee {
		name: String
		company: Company
	}
`

// setupEmployeeCompanyDB creates an in-memory DB with the Employee/Company schema
// and returns the two collection handles. The collections are returned as
// concrete *collection so tests can reach package-private methods without
// repeating the type assertion at every call site.
func setupEmployeeCompanyDB(t *testing.T) (*DB, *collection, *collection) {
	t.Helper()
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)

	_, err = db.AddCollection(ctx, employeeCompanySchema)
	require.NoError(t, err)

	empCol, err := db.GetCollectionByName(ctx, "Employee")
	require.NoError(t, err)

	companyCol, err := db.GetCollectionByName(ctx, "Company")
	require.NoError(t, err)

	empColImpl, ok := empCol.(*collection)
	require.True(t, ok)
	companyColImpl, ok := companyCol.(*collection)
	require.True(t, ok)

	return db, empColImpl, companyColImpl
}

// --- docExistsAndNotDeleted ---

func TestDocExistsAndNotDeleted_Exists(t *testing.T) {
	ctx := context.Background()
	db, _, companyCol := setupEmployeeCompanyDB(t)
	defer db.Close()

	doc, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Acme"}`), companyCol.Version())
	require.NoError(t, err)
	require.NoError(t, companyCol.AddDocument(ctx, doc))

	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(ctx, txn)

	primaryKey, err := companyCol.getPrimaryKeyFromDocID(txnCtx, doc.ID())
	require.NoError(t, err)

	exists, err := companyCol.docExistsAndNotDeleted(txnCtx, primaryKey)
	require.NoError(t, err)
	require.True(t, exists)
}

func TestDocExistsAndNotDeleted_NotFound(t *testing.T) {
	ctx := context.Background()
	db, _, companyCol := setupEmployeeCompanyDB(t)
	defer db.Close()

	// Create a doc object but do NOT save it — its primary key will not exist in the store.
	phantom, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Ghost"}`), companyCol.Version())
	require.NoError(t, err)

	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(ctx, txn)

	primaryKey, err := companyCol.getPrimaryKeyFromDocID(txnCtx, phantom.ID())
	require.NoError(t, err)

	exists, err := companyCol.docExistsAndNotDeleted(txnCtx, primaryKey)
	require.NoError(t, err)
	require.False(t, exists)
}

func TestDocExistsAndNotDeleted_SoftDeleted(t *testing.T) {
	ctx := context.Background()
	db, _, companyCol := setupEmployeeCompanyDB(t)
	defer db.Close()

	doc, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Acme"}`), companyCol.Version())
	require.NoError(t, err)
	require.NoError(t, companyCol.AddDocument(ctx, doc))

	_, err = companyCol.DeleteDocument(ctx, doc.ID())
	require.NoError(t, err)

	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(ctx, txn)

	primaryKey, err := companyCol.getPrimaryKeyFromDocID(txnCtx, doc.ID())
	require.NoError(t, err)

	exists, err := companyCol.docExistsAndNotDeleted(txnCtx, primaryKey)
	require.NoError(t, err)
	require.False(t, exists)
}

// --- validateRelationDocIDs ---

func TestValidateRelationDocIDs_ValidTarget_NoError(t *testing.T) {
	ctx := context.Background()
	db, empCol, companyCol := setupEmployeeCompanyDB(t)
	defer db.Close()

	companyDoc, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Acme"}`), companyCol.Version())
	require.NoError(t, err)
	require.NoError(t, companyCol.AddDocument(ctx, companyDoc))

	empDoc, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Alice"}`), empCol.Version())
	require.NoError(t, err)
	require.NoError(t, empDoc.Set(ctx, "_companyID", companyDoc.ID().String()))

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(ctx, txn)

	err = empCol.validateRelationDocIDs(txnCtx, empDoc)
	require.NoError(t, err)
}

func TestValidateRelationDocIDs_NonExistentTarget_Error(t *testing.T) {
	ctx := context.Background()
	db, empCol, companyCol := setupEmployeeCompanyDB(t)
	defer db.Close()

	// Create a company doc object that is never saved — its DocID will not exist in the store.
	phantomCompany, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Ghost Inc"}`), companyCol.Version())
	require.NoError(t, err)

	empDoc, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Alice"}`), empCol.Version())
	require.NoError(t, err)
	require.NoError(t, empDoc.Set(ctx, "_companyID", phantomCompany.ID().String()))

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(ctx, txn)

	err = empCol.validateRelationDocIDs(txnCtx, empDoc)
	require.ErrorContains(t, err, "relation target document not found")
}

func TestValidateRelationDocIDs_EmptyID_NoError(t *testing.T) {
	ctx := context.Background()
	db, empCol, _ := setupEmployeeCompanyDB(t)
	defer db.Close()

	// A doc whose _companyID is not set at all (zero value) — clearing a link is always valid.
	empDoc, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Alice"}`), empCol.Version())
	require.NoError(t, err)

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(ctx, txn)

	err = empCol.validateRelationDocIDs(txnCtx, empDoc)
	require.NoError(t, err)
}

func TestValidateRelationDocIDs_DeletedTarget_Error(t *testing.T) {
	ctx := context.Background()
	db, empCol, companyCol := setupEmployeeCompanyDB(t)
	defer db.Close()

	companyDoc, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Acme"}`), companyCol.Version())
	require.NoError(t, err)
	require.NoError(t, companyCol.AddDocument(ctx, companyDoc))
	_, err = companyCol.DeleteDocument(ctx, companyDoc.ID())
	require.NoError(t, err)

	empDoc, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Alice"}`), empCol.Version())
	require.NoError(t, err)
	require.NoError(t, empDoc.Set(ctx, "_companyID", companyDoc.ID().String()))

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(ctx, txn)

	err = empCol.validateRelationDocIDs(txnCtx, empDoc)
	require.ErrorContains(t, err, "relation target document not found")
}

func TestValidateRelationDocIDs_SkipContext_NoError(t *testing.T) {
	ctx := context.Background()
	db, empCol, companyCol := setupEmployeeCompanyDB(t)
	defer db.Close()

	// Phantom company — never saved.
	phantomCompany, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Ghost Inc"}`), companyCol.Version())
	require.NoError(t, err)

	empDoc, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Alice"}`), empCol.Version())
	require.NoError(t, err)
	require.NoError(t, empDoc.Set(ctx, "_companyID", phantomCompany.ID().String()))

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	defer txn.Discard()
	// skipRelationValidationContext suppresses all validation — used by backup import.
	txnCtx := InitContext(skipRelationValidationContext(ctx), txn)

	err = empCol.validateRelationDocIDs(txnCtx, empDoc)
	require.NoError(t, err)
}

func TestValidateRelationDocIDs_NonDirtyField_NoError(t *testing.T) {
	ctx := context.Background()
	db, empCol, companyCol := setupEmployeeCompanyDB(t)
	defer db.Close()

	// Create company, create employee linking to it.
	companyDoc, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Acme"}`), companyCol.Version())
	require.NoError(t, err)
	require.NoError(t, companyCol.AddDocument(ctx, companyDoc))

	// Include _companyID in the initial map so the DocID is computed from all fields.
	empDoc, err := client.NewDocFromMap(ctx, map[string]any{
		"name":       "Alice",
		"_companyID": companyDoc.ID().String(),
	}, empCol.Version())
	require.NoError(t, err)
	require.NoError(t, empCol.AddDocument(ctx, empDoc))

	// Delete the company. The employee now has a dangling _companyID in the store.
	_, err = companyCol.DeleteDocument(ctx, companyDoc.ID())
	require.NoError(t, err)

	// Retrieve the employee. Fields from GetDocument are NOT dirty.
	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(ctx, txn)

	retrieved, err := empCol.GetDocument(txnCtx, empDoc.ID())
	require.NoError(t, err)

	// validateRelationDocIDs skips non-dirty fields, so the dangling link is not an error.
	err = empCol.validateRelationDocIDs(txnCtx, retrieved)
	require.NoError(t, err)
}

// --- validateMergeRelationDocIDs ---

func TestValidateMergeRelationDocIDs_ValidTarget_NoError(t *testing.T) {
	ctx := context.Background()
	db, empCol, companyCol := setupEmployeeCompanyDB(t)
	defer db.Close()

	companyDoc, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Acme"}`), companyCol.Version())
	require.NoError(t, err)
	require.NoError(t, companyCol.AddDocument(ctx, companyDoc))

	empDoc, err := client.NewDocFromMap(ctx, map[string]any{
		"name":       "Alice",
		"_companyID": companyDoc.ID().String(),
	}, empCol.Version())
	require.NoError(t, err)
	require.NoError(t, empCol.AddDocument(ctx, empDoc))

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(ctx, txn)

	retrieved, err := empCol.GetDocument(txnCtx, empDoc.ID())
	require.NoError(t, err)

	err = empCol.validateMergeRelationDocIDs(txnCtx, retrieved)
	require.NoError(t, err)
}

func TestValidateMergeRelationDocIDs_MissingTarget_NoError(t *testing.T) {
	ctx := context.Background()
	db, empCol, companyCol := setupEmployeeCompanyDB(t)
	defer db.Close()

	// Phantom company: valid DocID format but never saved.
	phantomCompany, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Ghost Inc"}`), companyCol.Version())
	require.NoError(t, err)

	// Include _companyID in the initial map; skip write-path validation so we can store a dangling link.
	empDoc, err := client.NewDocFromMap(ctx, map[string]any{
		"name":       "Alice",
		"_companyID": phantomCompany.ID().String(),
	}, empCol.Version())
	require.NoError(t, err)
	require.NoError(t, empCol.AddDocument(skipRelationValidationContext(ctx), empDoc))

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(ctx, txn)

	retrieved, err := empCol.GetDocument(txnCtx, empDoc.ID())
	require.NoError(t, err)

	// Merge-path validation treats a missing target as a skip, not an error.
	err = empCol.validateMergeRelationDocIDs(txnCtx, retrieved)
	require.NoError(t, err)
}

func TestValidateMergeRelationDocIDs_DeletedTarget_NoError(t *testing.T) {
	ctx := context.Background()
	db, empCol, companyCol := setupEmployeeCompanyDB(t)
	defer db.Close()

	companyDoc, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Acme"}`), companyCol.Version())
	require.NoError(t, err)
	require.NoError(t, companyCol.AddDocument(ctx, companyDoc))

	empDoc, err := client.NewDocFromMap(ctx, map[string]any{
		"name":       "Alice",
		"_companyID": companyDoc.ID().String(),
	}, empCol.Version())
	require.NoError(t, err)
	require.NoError(t, empCol.AddDocument(ctx, empDoc))

	// Soft-delete the company after the employee was linked.
	_, err = companyCol.DeleteDocument(ctx, companyDoc.ID())
	require.NoError(t, err)

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(ctx, txn)

	retrieved, err := empCol.GetDocument(txnCtx, empDoc.ID())
	require.NoError(t, err)

	// A soft-deleted target is treated as "not found" on the merge path — skipped, not an error.
	err = empCol.validateMergeRelationDocIDs(txnCtx, retrieved)
	require.NoError(t, err)
}
