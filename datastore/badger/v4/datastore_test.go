// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package badger

import (
	"context"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testKey1   = ds.NewKey("testKey1")
	testValue1 = []byte("this is a test value 1")

	testKey2   = ds.NewKey("testKey2")
	testValue2 = []byte("this is a test value 2")

	testKey3   = ds.NewKey("testKey3")
	testValue3 = []byte("this is a test value 3")

	testKey4   = ds.NewKey("testKey4")
	testValue4 = []byte("this is a test value 4")

	testKey5   = ds.NewKey("testKey5")
	testValue5 = []byte("this is a test value 5")

	testKey6   = ds.NewKey("testKey6")
	testValue6 = []byte("this is a test value 6")
)

func newLoadedDatastore(ctx context.Context, t *testing.T) *Datastore {
	dir := t.TempDir()
	s, err := NewDatastore(dir, nil)
	require.NoError(t, err)
	s.Put(ctx, testKey1, testValue1)
	s.Put(ctx, testKey2, testValue2)
	return s
}

func TestNewDatastoreWithOptions(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	opt := DefaultOptions
	opt.GcInterval = time.Minute
	opt.GcSleep = 0

	s, err := NewDatastore(dir, &opt)
	require.NoError(t, err)

	s.Put(ctx, testKey1, testValue1)
	s.Put(ctx, testKey2, testValue2)
}

func TestNewBatch(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	b, err := s.Batch(ctx)
	require.NoError(t, err)
	assert.NotNil(t, b)
}

func TestBatchOperations(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	b, err := s.Batch(ctx)
	require.NoError(t, err)

	err = b.Delete(ctx, testKey1)
	require.NoError(t, err)

	err = b.Put(ctx, testKey3, testValue3)
	require.NoError(t, err)

	err = b.Commit(ctx)
	require.NoError(t, err)

	_, err = s.Get(ctx, testKey1)
	assert.ErrorIs(t, err, ds.ErrNotFound)

	resp, err := s.Get(ctx, testKey3)
	require.NoError(t, err)
	assert.Equal(t, testValue3, resp)
}

func TestBatchWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	err := s.Close()
	require.NoError(t, err)

	_, err = s.Batch(ctx)
	require.ErrorIs(t, err, ErrClosed)
}

func TestBatchPutWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	b, err := s.Batch(ctx)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	err = b.Put(ctx, testKey3, testValue3)
	require.ErrorIs(t, err, ErrClosed)
}

func TestBatchDeleteWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	b, err := s.Batch(ctx)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	err = b.Delete(ctx, testKey3)
	require.ErrorIs(t, err, ErrClosed)
}

func TestBatchCommitWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	b, err := s.Batch(ctx)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	err = b.Commit(ctx)
	require.ErrorIs(t, err, ErrClosed)
}

func TestBatchConsecutiveCommit(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	b, err := s.Batch(ctx)
	require.NoError(t, err)

	err = b.Commit(ctx)
	require.NoError(t, err)

	err = b.Commit(ctx)
	require.Equal(t, err.Error(), "Batch commit not permitted after finish")
}

func TestCollectGarbage(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	err := s.CollectGarbage(ctx)
	require.NoError(t, err)
}

func TestCollectGarbageWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	err := s.Close()
	require.NoError(t, err)

	err = s.CollectGarbage(ctx)
	require.ErrorIs(t, err, ErrClosed)
}

func TestCloseOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	err := s.Close()
	require.NoError(t, err)
}

func TestConsecutiveClose(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	err := s.Close()
	require.NoError(t, err)

	err = s.Close()
	require.ErrorIs(t, err, ErrClosed)
}

func TestGetOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	resp, err := s.Get(ctx, testKey1)
	require.NoError(t, err)
	require.Equal(t, testValue1, resp)
}

func TestGetOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
	err := s.Close()
	require.NoError(t, err)

	_, err = s.Get(ctx, testKey1)
	require.ErrorIs(t, err, ErrClosed)
}

func TestGetOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	_, err := s.Get(ctx, testKey3)
	require.ErrorIs(t, err, ds.ErrNotFound)
}

func TestDeleteOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	err := s.Delete(ctx, testKey1)
	require.NoError(t, err)

	_, err = s.Get(ctx, testKey1)
	require.ErrorIs(t, err, ds.ErrNotFound)
}

func TestDeleteOperation2(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	err := s.Put(ctx, testKey1, testValue1)
	require.NoError(t, err)

	err = s.Delete(ctx, testKey1)
	require.NoError(t, err)

	_, err = s.Get(ctx, testKey1)
	require.ErrorIs(t, err, ds.ErrNotFound)
}

func TestDeleteOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	err := s.Close()
	require.NoError(t, err)

	err = s.Delete(ctx, testKey1)
	require.ErrorIs(t, err, badger.ErrBlockedWrites)
}

func TestGetSizeOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	resp, err := s.GetSize(ctx, testKey1)
	require.NoError(t, err)
	require.Equal(t, len(testValue1), resp)
}

func TestGetSizeOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	_, err := s.GetSize(ctx, testKey3)
	require.ErrorIs(t, err, ds.ErrNotFound)
}

func TestGetSizeOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	err := s.Close()
	require.NoError(t, err)

	_, err = s.GetSize(ctx, testKey3)
	require.ErrorIs(t, err, ErrClosed)
}

func TestHasOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	resp, err := s.Has(ctx, testKey1)
	require.NoError(t, err)
	require.Equal(t, true, resp)
}

func TestHasOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	resp, err := s.Has(ctx, testKey3)
	require.NoError(t, err)
	require.Equal(t, false, resp)
}

func TestHasOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	err := s.Close()
	require.NoError(t, err)

	_, err = s.Has(ctx, testKey3)
	require.ErrorIs(t, err, ErrClosed)
}

func TestPutOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	err := s.Put(ctx, testKey3, testValue3)
	require.NoError(t, err)

	resp, err := s.Get(ctx, testKey3)
	require.NoError(t, err)
	require.Equal(t, testValue3, resp)
}

func TestPutOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	err := s.Close()
	require.NoError(t, err)

	err = s.Put(ctx, testKey3, testValue3)
	require.ErrorIs(t, err, ErrClosed)
}

func TestQueryOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	results, err := s.Query(ctx, dsq.Query{
		Limit:  1,
		Offset: 1,
	})
	require.NoError(t, err)

	result, _ := results.NextSync()

	require.Equal(t, testKey2.String(), result.Entry.Key)
	require.Equal(t, testValue2, result.Entry.Value)
}

func TestQueryOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	err := s.Close()
	require.NoError(t, err)

	_, err = s.Query(ctx, dsq.Query{
		Limit:  1,
		Offset: 1,
	})
	require.ErrorIs(t, err, ErrClosed)
}

func TestDiskUsage(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	size, err := s.DiskUsage(ctx)
	require.NoError(t, err)
	require.Equal(t, size, uint64(0))
}

func TestDiskUsageWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	err := s.Close()
	require.NoError(t, err)

	_, err = s.DiskUsage(ctx)
	require.ErrorIs(t, err, ErrClosed)
}

func TestSync(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
	err := s.Sync(ctx, testKey1)
	require.NoError(t, err)
}

func TestSyncWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	err := s.Close()
	require.NoError(t, err)

	err = s.Sync(ctx, testKey1)
	require.ErrorIs(t, err, ErrClosed)
}

func TestNewTransaction(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
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
	s := newLoadedDatastore(ctx, t)

	err := s.Close()
	require.NoError(t, err)
	_, err = s.NewTransaction(ctx, false)
	require.ErrorIs(t, err, ErrClosed)
}

func TestTxnDeleteOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	err = tx.Delete(ctx, testKey1)
	require.ErrorIs(t, err, ErrClosed)
}

func TestTxnGetOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
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
	s := newLoadedDatastore(ctx, t)

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	_, err = tx.Get(ctx, testKey1)
	require.ErrorIs(t, err, ErrClosed)
}

func TestTxnGetOperationAfterPut(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
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
	s := newLoadedDatastore(ctx, t)
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

func TestTxnDeleteOperationReadOnly(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, true)
	require.NoError(t, err)

	err = tx.Delete(ctx, testKey1)
	require.ErrorIs(t, err, badger.ErrReadOnlyTxn)
}

func TestTxnGetOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
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
	s := newLoadedDatastore(ctx, t)
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
	require.ErrorIs(t, err, badger.ErrDiscardedTxn)

	_, err = s.Get(ctx, testKey1)
	require.ErrorIs(t, err, ds.ErrNotFound)
}

func TestTxnGetSizeOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
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
	s := newLoadedDatastore(ctx, t)

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	_, err = tx.GetSize(ctx, testKey1)
	require.ErrorIs(t, err, ErrClosed)
}

func TestTxnGetSizeOperationWithDiscardedTxn(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	tx.Discard(ctx)

	_, err = tx.GetSize(ctx, testKey1)
	require.ErrorIs(t, err, badger.ErrDiscardedTxn)
}

func TestTxnGetSizeOfterPutOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
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
	require.Equal(t, len(testValue3), resp)
}

func TestTxnGetSizeOperationAfterDelete(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
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
	s := newLoadedDatastore(ctx, t)
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
	s := newLoadedDatastore(ctx, t)
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
	s := newLoadedDatastore(ctx, t)

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	_, err = tx.Has(ctx, testKey1)
	require.ErrorIs(t, err, ErrClosed)
}

func TestTxnHasOperationWithDiscardedTxn(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	tx.Discard(ctx)

	_, err = tx.Has(ctx, testKey1)
	require.ErrorIs(t, err, badger.ErrDiscardedTxn)
}

func TestTxnHasOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
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
	s := newLoadedDatastore(ctx, t)
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
	s := newLoadedDatastore(ctx, t)
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
	s := newLoadedDatastore(ctx, t)
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
	s := newLoadedDatastore(ctx, t)

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	err = tx.Put(ctx, testKey1, testValue1)
	require.ErrorIs(t, err, ErrClosed)
}

func TestTxnPutOperationWithDiscardedTxn(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	tx.Discard(ctx)

	err = tx.Put(ctx, testKey1, testValue1)
	require.ErrorIs(t, err, badger.ErrDiscardedTxn)
}

func TestTxnPutAndCommitOperationReadOnly(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, true)
	require.NoError(t, err)

	err = tx.Put(ctx, testKey3, testValue3)
	require.ErrorIs(t, err, badger.ErrReadOnlyTxn)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	_, err = s.Get(ctx, testKey3)
	require.ErrorIs(t, err, ds.ErrNotFound)
}

func TestTxnPutOperationReadOnly(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()
	tx, err := s.NewTransaction(ctx, true)
	require.NoError(t, err)

	err = tx.Put(ctx, testKey3, testValue3)
	require.ErrorIs(t, err, badger.ErrReadOnlyTxn)
}

func TestTxnQueryOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
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
	s := newLoadedDatastore(ctx, t)

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

func TestTxnQueryOperationInTwoConcurentTxn(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
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
	s := newLoadedDatastore(ctx, t)
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
	s := newLoadedDatastore(ctx, t)
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

func TestTxnWithConflict(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
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

func TestTxnWithConflictAfterDelete(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	tx2, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	_, err = tx.GetSize(ctx, testKey2)
	require.NoError(t, err)

	err = tx.Put(ctx, testKey2, testValue3)
	require.NoError(t, err)

	err = tx2.Delete(ctx, testKey2)
	require.NoError(t, err)

	err = tx2.Commit(ctx)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.ErrorIs(t, err, ErrTxnConflict)
}

func TestTxnWithNoConflictAfterGet(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)
	defer func() {
		err := s.Close()
		require.NoError(t, err)
	}()

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	tx2, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	_, err = tx.Get(ctx, testKey2)
	require.NoError(t, err)

	err = tx2.Put(ctx, testKey2, testValue3)
	require.NoError(t, err)

	err = tx2.Commit(ctx)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)
}

func TestTxnCommitWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.ErrorIs(t, err, ErrClosed)
}

func TestTxnClose(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.ErrorIs(t, err, ErrClosed)
}

func TestTxnDiscardWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx, t)

	tx, err := s.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	tx.Discard(ctx)
}
