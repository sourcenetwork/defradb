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
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
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

func newLoadedDatastore(ctx context.Context) *Datastore {
	s := NewDatastore(ctx)
	v := s.nextVersion()
	s.values.Set(dsItem{
		key:     testKey1.String(),
		val:     testValue1,
		version: v,
	})
	v = s.nextVersion()
	s.values.Set(dsItem{
		key:     testKey2.String(),
		val:     testValue2,
		version: v,
	})
	return s
}

func TestNewDatastore(t *testing.T) {
	ctx := context.Background()
	s := NewDatastore(ctx)
	require.NotNil(t, s)
}

func TestCloseOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	err := s.Close()
	require.NoError(t, err)
}

func TestConsecutiveClose(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	err := s.Close()
	require.NoError(t, err)

	err = s.Close()
	require.ErrorIs(t, err, ErrClosed)
}

func TestCloseThroughContext(t *testing.T) {
	ctx := context.Background()
	newCtx, cancel := context.WithCancel(ctx)
	s := newLoadedDatastore(newCtx)

	cancel()

	time.Sleep(time.Millisecond * 10)

	err := s.Close()
	require.ErrorIs(t, err, ErrClosed)
}

func TestGetOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	resp, err := s.Get(ctx, testKey1)
	require.NoError(t, err)
	require.Equal(t, testValue1, resp)
}

func TestGetOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	err := s.Close()
	require.NoError(t, err)

	_, err = s.Get(ctx, testKey1)
	require.ErrorIs(t, err, ErrClosed)
}

func TestGetOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	_, err := s.Get(ctx, testKey3)
	require.ErrorIs(t, err, ds.ErrNotFound)
}

func TestDeleteOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	err := s.Delete(ctx, testKey1)
	require.NoError(t, err)

	_, err = s.Get(ctx, testKey1)
	require.ErrorIs(t, err, ds.ErrNotFound)
}

func TestDeleteOperation2(t *testing.T) {
	ctx := context.Background()
	s := NewDatastore(ctx)

	err := s.Put(ctx, testKey1, testValue1)
	require.NoError(t, err)

	err = s.Delete(ctx, testKey1)
	require.NoError(t, err)

	_, err = s.Get(ctx, testKey1)
	require.ErrorIs(t, err, ds.ErrNotFound)
}

func TestDeleteOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	err := s.Close()
	require.NoError(t, err)

	err = s.Delete(ctx, testKey1)
	require.ErrorIs(t, err, ErrClosed)
}

func TestGetSizeOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	resp, err := s.GetSize(ctx, testKey1)
	require.NoError(t, err)
	require.Equal(t, len(testValue1), resp)
}

func TestGetSizeOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	_, err := s.GetSize(ctx, testKey3)
	require.ErrorIs(t, err, ds.ErrNotFound)
}

func TestGetSizeOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	err := s.Close()
	require.NoError(t, err)

	_, err = s.GetSize(ctx, testKey3)
	require.ErrorIs(t, err, ErrClosed)
}

func TestHasOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	resp, err := s.Has(ctx, testKey1)
	require.NoError(t, err)
	require.Equal(t, true, resp)
}

func TestHasOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	resp, err := s.Has(ctx, testKey3)
	require.NoError(t, err)
	require.Equal(t, false, resp)
}

func TestHasOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	err := s.Close()
	require.NoError(t, err)

	_, err = s.Has(ctx, testKey3)
	require.ErrorIs(t, err, ErrClosed)
}

func TestPutOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	err := s.Put(ctx, testKey3, testValue3)
	require.NoError(t, err)

	resp, err := s.Get(ctx, testKey3)
	require.NoError(t, err)
	require.Equal(t, testValue3, resp)
}

func TestPutOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	err := s.Close()
	require.NoError(t, err)

	err = s.Put(ctx, testKey3, testValue3)
	require.ErrorIs(t, err, ErrClosed)
}

func TestQueryOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

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
	s := newLoadedDatastore(ctx)

	err := s.Close()
	require.NoError(t, err)

	_, err = s.Query(ctx, dsq.Query{
		Limit:  1,
		Offset: 1,
	})
	require.ErrorIs(t, err, ErrClosed)
}

func TestQueryOperationWithAddedItems(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	err := s.Put(ctx, testKey3, testValue3)
	require.NoError(t, err)
	err = s.Put(ctx, testKey4, testValue4)
	require.NoError(t, err)

	err = s.Put(ctx, testKey5, testValue5)
	require.NoError(t, err)

	err = s.Delete(ctx, testKey2)
	require.NoError(t, err)

	err = s.Put(ctx, testKey2, testValue2)
	require.NoError(t, err)

	err = s.Delete(ctx, testKey1)
	require.NoError(t, err)

	results, err := s.Query(ctx, dsq.Query{})
	require.NoError(t, err)
	entries, err := results.Rest()
	require.NoError(t, err)
	expectedResults := []dsq.Entry{
		{
			Key:   testKey2.String(),
			Value: testValue2,
			Size:  len(testValue2),
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
	}
	require.Equal(t, expectedResults, entries)
}

func TestConcurrentWrite(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	wg := &sync.WaitGroup{}

	for i := 1; i <= 1000; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, num int) {
			_ = s.Put(ctx, ds.NewKey(fmt.Sprintf("testKey%d", num)), []byte(fmt.Sprintf("this is a test value %d", num)))
			wg.Done()
		}(wg, i)
	}
	wg.Wait()
	resp, err := s.Get(ctx, ds.NewKey("testKey3"))
	require.NoError(t, err)
	require.Equal(t, []byte("this is a test value 3"), resp)
}

func TestSyncOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	err := s.Sync(ctx, testKey1)
	require.NoError(t, err)
}

func TestSyncOperationWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	err := s.Close()
	require.NoError(t, err)

	err = s.Sync(ctx, testKey1)
	require.ErrorIs(t, err, ErrClosed)
}

func TestPurge(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	err := s.Put(ctx, testKey1, testValue2)
	require.NoError(t, err)
	err = s.Put(ctx, testKey2, testValue3)
	require.NoError(t, err)

	iter := s.values.Iter()
	results := []dsq.Entry{}
	for iter.Next() {
		results = append(results, dsq.Entry{
			Key:   iter.Item().key,
			Value: iter.Item().val,
			Size:  len(iter.Item().val),
		})
	}
	iter.Release()

	expectedResults := []dsq.Entry{
		{
			Key:   testKey1.String(),
			Value: testValue1,
			Size:  len(testValue1),
		},
		{
			Key:   testKey1.String(),
			Value: testValue2,
			Size:  len(testValue2),
		},
		{
			Key:   testKey2.String(),
			Value: testValue2,
			Size:  len(testValue2),
		},
		{
			Key:   testKey2.String(),
			Value: testValue3,
			Size:  len(testValue3),
		},
	}
	require.Equal(t, expectedResults, results)

	s.executePurge(ctx)

	iter = s.values.Iter()
	results = []dsq.Entry{}
	for iter.Next() {
		results = append(results, dsq.Entry{
			Key:   iter.Item().key,
			Value: iter.Item().val,
			Size:  len(iter.Item().val),
		})
	}
	iter.Release()

	expectedResults = []dsq.Entry{
		{
			Key:   testKey1.String(),
			Value: testValue2,
			Size:  len(testValue2),
		},
		{
			Key:   testKey2.String(),
			Value: testValue3,
			Size:  len(testValue3),
		},
	}
	require.Equal(t, expectedResults, results)
}

func TestPurgeBatching(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	for j := 0; j < 10; j++ {
		for i := 1; i <= 1000; i++ {
			err := s.Put(ctx, ds.NewKey("test"), []byte(fmt.Sprintf("%d", i+(j*1000))))
			require.NoError(t, err)
		}
	}

	s.executePurge(ctx)

	resp, err := s.Get(ctx, ds.NewKey("test"))
	require.NoError(t, err)

	val, err := strconv.Atoi(string(resp))
	require.NoError(t, err)

	require.GreaterOrEqual(t, val, 9000)
}

func TestPurgeWithOlderInFlightTxn(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)
	s.inFlightTxn.Set(dsTxn{
		dsVersion:  s.getVersion(),
		txnVersion: s.getVersion() + 1,
		expiresAt:  time.Now(),
	})

	err := s.Put(ctx, testKey4, testValue4)
	require.NoError(t, err)

	s.executePurge(ctx)
}

func TestClearOldFlightTransactions(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	s.inFlightTxn.Set(dsTxn{
		dsVersion:  s.getVersion(),
		txnVersion: s.getVersion() + 1,
		// Ensure expiresAt is before the value returned from the later call in `clearOldInFlightTxn`,
		// in windows in particular it seems that the two `time.Now` calls can return the same value
		expiresAt: time.Now().Add(-1 * time.Minute),
	})

	require.Equal(t, 1, s.inFlightTxn.Len())

	s.clearOldInFlightTxn(ctx)

	require.Equal(t, 0, s.inFlightTxn.Len())
}
