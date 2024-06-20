// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package memory

import (
	"context"
	"testing"

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/btree"
)

func TestNewTransaction(t *testing.T) {
	ctx := context.Background()
	s := NewDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, false)
	require.NotNil(t, tx)
	require.NoError(t, err)
}

func TestNewTransactionWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := NewDatastore(ctx)

	err := s.Close()
	require.NoError(t, err)
	_, err = s.NewTransaction(ctx, false)
	require.ErrorIs(t, err, ErrClosed)
}

func TestTxnDeleteOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	err = tx.Delete(ctx, testKey1)
	require.ErrorIs(t, err, ErrClosed)
}

func TestTxnDeleteOperationWithDiscardedTxn(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	tx.Discard(ctx)

	err = tx.Delete(ctx, testKey1)
	require.ErrorIs(t, err, ErrTxnDiscarded)
}

func TestTxnGetOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	resp, err := tx.Get(ctx, testKey1)
	require.NoError(t, err)
	require.Equal(t, testValue1, resp)
}

func TestTxnGetOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	_, err = tx.Get(ctx, testKey1)
	require.ErrorIs(t, err, ErrClosed)
}

func TestTxnGetOperationAfterPut(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = tx.Put(ctx, testKey3, testValue3)
	require.NoError(t, err)

	resp, err := tx.Get(ctx, testKey3)
	require.NoError(t, err)
	require.Equal(t, testValue3, resp)
}

func TestTxnGetOperationAfterDelete(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = tx.Delete(ctx, testKey1)
	require.NoError(t, err)

	_, err = tx.Get(ctx, testKey1)
	require.ErrorIs(t, err, ds.ErrNotFound)
}

func TestTxnGetOperationAfterDeleteReadOnly(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, true)
	require.NoError(t, err)

	err = tx.Delete(ctx, testKey1)
	require.ErrorIs(t, err, ErrReadOnlyTxn)
}

func TestTxnGetOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	_, err = tx.Get(ctx, testKey3)
	require.ErrorIs(t, err, ds.ErrNotFound)
}

func TestTxnDeleteAndCommitOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = tx.Delete(ctx, testKey1)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	_, err = tx.Get(ctx, testKey1)
	require.ErrorIs(t, err, ErrTxnDiscarded)

	_, err = s.Get(ctx, testKey1)
	require.ErrorIs(t, err, ds.ErrNotFound)
}

func TestTxnGetSizeOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	resp, err := tx.GetSize(ctx, testKey1)
	require.NoError(t, err)
	require.Equal(t, len(testValue1), resp)
}

func TestTxnGetSizeOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	_, err = tx.GetSize(ctx, testKey1)
	require.ErrorIs(t, err, ErrClosed)
}

func TestTxnGetSizeOperationWithDiscardedTxn(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	tx.Discard(ctx)

	_, err = tx.GetSize(ctx, testKey1)
	require.ErrorIs(t, err, ErrTxnDiscarded)
}

func TestTxnGetSizeOfterPutOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = tx.Put(ctx, testKey3, testValue3)
	require.NoError(t, err)

	resp, err := tx.GetSize(ctx, testKey3)
	require.NoError(t, err)
	require.Equal(t, len(testValue1), resp)
}

func TestTxnGetSizeOperationAfterDelete(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = tx.Delete(ctx, testKey1)
	require.NoError(t, err)

	_, err = tx.GetSize(ctx, testKey1)
	require.ErrorIs(t, err, ds.ErrNotFound)
}

func TestTxnGetSizeOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	_, err = tx.GetSize(ctx, testKey3)
	require.ErrorIs(t, err, ds.ErrNotFound)
}

func TestTxnHasOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	resp, err := tx.Has(ctx, testKey1)
	require.NoError(t, err)
	require.Equal(t, true, resp)
}

func TestTxnHasOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	_, err = tx.Has(ctx, testKey1)
	require.ErrorIs(t, err, ErrClosed)
}

func TestTxnHasOperationWithDiscardedTxn(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	tx.Discard(ctx)

	_, err = tx.Has(ctx, testKey1)
	require.ErrorIs(t, err, ErrTxnDiscarded)
}

func TestTxnHasOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	resp, err := tx.Has(ctx, testKey3)
	require.NoError(t, err)
	require.Equal(t, false, resp)
}

func TestTxnHasOfterPutOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = tx.Put(ctx, testKey3, testValue3)
	require.NoError(t, err)

	resp, err := tx.Has(ctx, testKey3)
	require.NoError(t, err)
	require.Equal(t, true, resp)
}

func TestTxnHasOperationAfterDelete(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = tx.Delete(ctx, testKey1)
	require.NoError(t, err)

	resp, err := tx.Has(ctx, testKey1)
	require.NoError(t, err)
	require.Equal(t, false, resp)
}

func TestTxnPutAndCommitOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = tx.Put(ctx, testKey3, testValue3)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	resp, err := s.Has(ctx, testKey3)
	require.NoError(t, err)
	require.Equal(t, true, resp)
}

func TestTxnPutOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	err = tx.Put(ctx, testKey1, testValue1)
	require.ErrorIs(t, err, ErrClosed)
}

func TestTxnPutOperationWithDiscardedTxn(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	tx.Discard(ctx)

	err = tx.Put(ctx, testKey1, testValue1)
	require.ErrorIs(t, err, ErrTxnDiscarded)
}

func TestTxnPutAndCommitOperationReadOnly(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, true)
	require.NoError(t, err)

	err = tx.Put(ctx, testKey3, testValue3)
	require.ErrorIs(t, err, ErrReadOnlyTxn)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	_, err = s.Get(ctx, testKey3)
	require.ErrorIs(t, err, ds.ErrNotFound)
}

func TestTxnPutOperationReadOnly(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, true)
	require.NoError(t, err)

	err = tx.Put(ctx, testKey3, testValue3)
	require.ErrorIs(t, err, ErrReadOnlyTxn)
}

func TestTxnQueryOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = s.Put(ctx, testKey4, testValue4)
	require.NoError(t, err)

	err = tx.Put(ctx, testKey3, testValue3)
	require.NoError(t, err)

	err = tx.Put(ctx, testKey2, testValue3)
	require.NoError(t, err)

	err = tx.Delete(ctx, testKey1)
	require.NoError(t, err)

	err = tx.Delete(ctx, testKey5)
	require.NoError(t, err)

	err = tx.Put(ctx, testKey6, testValue6)
	require.NoError(t, err)

	results, err := tx.Query(ctx, dsq.Query{
		Limit:  1,
		Offset: 1,
	})
	require.NoError(t, err)

	result, _ := results.NextSync()

	require.Equal(t, testKey3.String(), result.Entry.Key)
	require.Equal(t, testValue3, result.Entry.Value)
}

func TestTxnQueryOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	_, err = tx.Query(ctx, dsq.Query{
		Limit:  1,
		Offset: 1,
	})
	require.ErrorIs(t, err, ErrClosed)
}

func TestTxnQueryOperationWithDiscardedTxn(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	tx.Discard(ctx)

	_, err = tx.Query(ctx, dsq.Query{
		Limit:  1,
		Offset: 1,
	})
	require.ErrorIs(t, err, ErrTxnDiscarded)
}

func TestTxnQueryOperationInTwoConcurentTxn(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)
	defer tx.Discard(ctx)

	tx2, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)
	defer tx.Discard(ctx)

	err = tx.Put(ctx, testKey1, testValue3)
	require.NoError(t, err)

	result, err := tx.Get(ctx, testKey1)
	require.NoError(t, err)

	require.Equal(t, testValue3, result)

	results, err := tx2.Query(ctx, dsq.Query{})
	require.NoError(t, err)
	entries, err := results.Rest()
	require.NoError(t, err)
	expectedResults := []dsq.Entry{
		{
			Key:   testKey1.String(),
			Value: testValue1,
			Size:  len(testValue1),
		},
		{
			Key:   testKey2.String(),
			Value: testValue2,
			Size:  len(testValue2),
		},
	}

	require.Equal(t, expectedResults, entries)
}

func TestTxnQueryOperationWithAddedItems(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()

	err := s.Put(ctx, testKey3, testValue3)
	require.NoError(t, err)
	err = s.Put(ctx, testKey4, testValue4)
	require.NoError(t, err)

	err = s.Put(ctx, testKey5, testValue5)
	require.NoError(t, err)

	err = s.Delete(ctx, testKey2)
	require.NoError(t, err)

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = tx.Put(ctx, testKey6, testValue6)
	require.NoError(t, err)

	err = tx.Put(ctx, testKey2, testValue3)
	require.NoError(t, err)

	err = tx.Delete(ctx, testKey1)
	require.NoError(t, err)

	results, err := tx.Query(ctx, dsq.Query{})
	require.NoError(t, err)
	entries, err := results.Rest()
	require.NoError(t, err)
	expectedResults := []dsq.Entry{
		{
			Key:   testKey2.String(),
			Value: testValue3,
			Size:  len(testValue3),
		},
		{
			Key:   testKey3.String(),
			Value: testValue3,
			Size:  len(testValue3),
		},
		{
			Key:   testKey4.String(),
			Value: testValue4,
			Size:  len(testValue4),
		},
		{
			Key:   testKey5.String(),
			Value: testValue5,
			Size:  len(testValue5),
		},
		{
			Key:   testKey6.String(),
			Value: testValue6,
			Size:  len(testValue6),
		},
	}
	require.Equal(t, expectedResults, entries)
}

func TestTxnQueryWithOnlyOneOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = tx.Put(ctx, testKey4, testValue4)
	require.NoError(t, err)

	results, err := tx.Query(ctx, dsq.Query{})
	require.NoError(t, err)

	_, _ = results.NextSync()
	_, _ = results.NextSync()
	result, _ := results.NextSync()

	require.Equal(t, testKey4.String(), result.Entry.Key)
	require.Equal(t, testValue4, result.Entry.Value)
}

func TestTxnDiscardOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	v := s.nextVersion()
	tx := &basicTxn{
		ops:       btree.NewBTreeG(byKeys),
		ds:        s,
		readOnly:  false,
		dsVersion: &v,
	}

	err := tx.Put(ctx, testKey3, testValue3)
	require.NoError(t, err)

	require.Equal(t, 1, tx.ops.Len())

	tx.Discard(ctx)
	require.Equal(t, 0, tx.ops.Len())
}

func TestTxnWithConflict(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	tx2, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	_, err = tx.GetSize(ctx, testKey3)
	require.ErrorIs(t, err, ds.ErrNotFound)

	err = tx.Put(ctx, testKey3, testValue3)
	require.NoError(t, err)

	err = tx2.Put(ctx, testKey3, testValue4)
	require.NoError(t, err)

	err = tx2.Commit(ctx)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.ErrorIs(t, err, ErrTxnConflict)
}

func TestTxnWithNoConflictAfterDelete(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()

	tx := s.newTransaction(false)

	tx2 := s.newTransaction(false)

	err := tx.Put(ctx, testKey2, testValue3)
	require.NoError(t, err)

	err = tx2.Delete(ctx, testKey2)
	require.NoError(t, err)

	err = tx2.Commit(ctx)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)
}

func TestTxnWithConflictAfterGet(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()

	tx := s.newTransaction(false)

	tx2 := s.newTransaction(false)

	_, err := tx.Get(ctx, testKey2)
	require.NoError(t, err)

	err = tx2.Put(ctx, testKey2, testValue3)
	require.NoError(t, err)

	err = tx2.Commit(ctx)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.ErrorIs(t, err, ErrTxnConflict)
}

func TestTxnCommitWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.ErrorIs(t, err, ErrClosed)
}

func TestTxnCommitWithDiscardedTxn(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	tx.Discard(ctx)

	err = tx.Commit(ctx)
	require.ErrorIs(t, err, ErrTxnDiscarded)
}
