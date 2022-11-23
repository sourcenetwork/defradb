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
	"sync"
	"sync/atomic"
	"testing"

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/stretchr/testify/assert"
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
	v := atomic.AddUint64(s.version, 1)
	s.values.Set(item{
		key:     testKey1.String(),
		val:     testValue1,
		version: v,
	})
	v++
	s.values.Set(item{
		key:     testKey2.String(),
		val:     testValue2,
		version: v,
	})
	return s
}

func TestNewDatastore(t *testing.T) {
	ctx := context.Background()
	s := NewDatastore(ctx)
	assert.NotNil(t, s)
}

func TestGetOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	resp, err := s.Get(ctx, testKey1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, testValue1, resp)
}

func TestGetOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	_, err := s.Get(ctx, testKey3)
	assert.ErrorIs(t, err, ds.ErrNotFound)
}

func TestDeleteOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	err := s.Delete(ctx, testKey1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Get(ctx, testKey1)
	assert.ErrorIs(t, err, ds.ErrNotFound)
}

func TestGetSizeOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	resp, err := s.GetSize(ctx, testKey1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(testValue1), resp)
}

func TestGetSizeOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	_, err := s.GetSize(ctx, testKey3)
	assert.ErrorIs(t, err, ds.ErrNotFound)
}

func TestHasOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	resp, err := s.Has(ctx, testKey1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, resp)
}

func TestHasOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	resp, err := s.Has(ctx, testKey3)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, false, resp)
}

func TestPutOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	err := s.Put(ctx, testKey3, testValue3)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := s.Get(ctx, testKey3)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, testValue3, resp)
}

func TestQueryOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	results, err := s.Query(ctx, dsq.Query{
		Limit:  1,
		Offset: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, _ := results.NextSync()

	assert.Equal(t, testKey2.String(), result.Entry.Key)
	assert.Equal(t, testValue2, result.Entry.Value)
}

func TestQueryOperationWithAddedItems(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	err := s.Put(ctx, testKey3, testValue3)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Put(ctx, testKey4, testValue4)
	if err != nil {
		t.Fatal(err)
	}

	err = s.Put(ctx, testKey5, testValue5)
	if err != nil {
		t.Fatal(err)
	}

	err = s.Delete(ctx, testKey2)
	if err != nil {
		t.Fatal(err)
	}

	err = s.Put(ctx, testKey2, testValue2)
	if err != nil {
		t.Fatal(err)
	}

	err = s.Delete(ctx, testKey1)
	if err != nil {
		t.Fatal(err)
	}

	results, err := s.Query(ctx, dsq.Query{})
	if err != nil {
		t.Fatal(err)
	}
	entries, err := results.Rest()
	if err != nil {
		t.Fatal(err)
	}
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
	assert.Equal(t, expectedResults, entries)
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
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []byte("this is a test value 3"), resp)
}

func TestCloseOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	err := s.Close()
	assert.NoError(t, err)
}

func TestSyncOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	err := s.Sync(ctx, testKey1)
	assert.NoError(t, err)
}

func TestCompressor(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore(ctx)

	err := s.Put(ctx, testKey1, testValue2)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Put(ctx, testKey2, testValue3)
	if err != nil {
		t.Fatal(err)
	}

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
	assert.Equal(t, expectedResults, results)

	s.smash(ctx)

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
	assert.Equal(t, expectedResults, results)
}
