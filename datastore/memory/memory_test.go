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

	testKey5 = ds.NewKey("testKey5")
)

func newLoadedStore() *Store {
	return &Store{
		values: map[ds.Key][]byte{
			testKey1: testValue1,
			testKey2: testValue2,
		},
	}
}

func TestNewStore(t *testing.T) {
	s := NewStore()
	assert.NotNil(t, s)
}

func TestGetOperation(t *testing.T) {
	s := newLoadedStore()

	ctx := context.Background()

	resp, err := s.Get(ctx, testKey1)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, testValue1, resp)
}

func TestGetOperationNotFound(t *testing.T) {
	s := newLoadedStore()

	ctx := context.Background()

	_, err := s.Get(ctx, testKey3)
	assert.ErrorIs(t, err, ds.ErrNotFound)
}

func TestDeleteOperation(t *testing.T) {
	s := newLoadedStore()

	ctx := context.Background()

	err := s.Delete(ctx, testKey1)
	if err != nil {
		t.Error(err)
	}

	_, err = s.Get(ctx, testKey1)
	assert.ErrorIs(t, err, ds.ErrNotFound)
}

func TestGetSizeOperation(t *testing.T) {
	s := newLoadedStore()

	ctx := context.Background()

	resp, err := s.GetSize(ctx, testKey1)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(testValue1), resp)
}

func TestGetSizeOperationNotFound(t *testing.T) {
	s := newLoadedStore()

	ctx := context.Background()

	_, err := s.GetSize(ctx, testKey3)
	assert.ErrorIs(t, err, ds.ErrNotFound)
}

func TestHasOperation(t *testing.T) {
	s := newLoadedStore()

	ctx := context.Background()

	resp, err := s.Has(ctx, testKey1)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, true, resp)
}

func TestHasOperationNotFound(t *testing.T) {
	s := newLoadedStore()

	ctx := context.Background()

	resp, err := s.Has(ctx, testKey3)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, false, resp)
}

func TestPutOperation(t *testing.T) {
	s := newLoadedStore()

	ctx := context.Background()

	err := s.Put(ctx, testKey3, testValue3)
	if err != nil {
		t.Error(err)
	}

	resp, err := s.Get(ctx, testKey3)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, testValue3, resp)
}

func TestQueryOperation(t *testing.T) {
	s := newLoadedStore()

	ctx := context.Background()

	results, err := s.Query(ctx, dsq.Query{
		Limit:  1,
		Offset: 1,
	})
	if err != nil {
		t.Error(err)
	}

	result, _ := results.NextSync()

	assert.Equal(t, testKey2.String(), result.Entry.Key)
	assert.Equal(t, testValue2, result.Entry.Value)
}

func TestCloseOperationNotFound(t *testing.T) {
	s := newLoadedStore()

	err := s.Close()
	assert.NoError(t, err)
}

func TestSyncOperationNotFound(t *testing.T) {
	s := newLoadedStore()

	ctx := context.Background()

	err := s.Sync(ctx, testKey1)
	assert.NoError(t, err)
}
