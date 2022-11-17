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
	"github.com/tidwall/btree"
)

func TestNewTransaction(t *testing.T) {
	ctx := context.Background()
	s := NewDatastore()
	tx, err := s.NewTransaction(ctx, false)
	assert.NotNil(t, tx)
	assert.NoError(t, err)
}

func TestTxnGetOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, false)
	if err != nil {
		t.Error(err)
	}

	resp, err := tx.Get(ctx, testKey1)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, testValue1, resp)
}

func TestTxnGetOperationAfterPut(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, false)
	if err != nil {
		t.Error(err)
	}

	err = tx.Put(ctx, testKey3, testValue3)
	if err != nil {
		t.Error(err)
	}

	resp, err := tx.Get(ctx, testKey3)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, testValue3, resp)
}

func TestTxnGetOperationAfterDelete(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, false)
	if err != nil {
		t.Error(err)
	}

	err = tx.Delete(ctx, testKey1)
	if err != nil {
		t.Error(err)
	}

	_, err = tx.Get(ctx, testKey1)
	assert.ErrorIs(t, err, ds.ErrNotFound)
}

func TestTxnGetOperationAfterDeleteReadOnly(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, true)
	if err != nil {
		t.Error(err)
	}

	err = tx.Delete(ctx, testKey1)
	assert.ErrorIs(t, err, ErrReadOnlyTxn)
}

func TestTxnGetOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, false)
	if err != nil {
		t.Error(err)
	}

	_, err = tx.Get(ctx, testKey3)
	assert.ErrorIs(t, err, ds.ErrNotFound)
}

func TestTxnDeleteAndCommitOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, false)
	if err != nil {
		t.Error(err)
	}

	err = tx.Delete(ctx, testKey1)
	if err != nil {
		t.Error(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		t.Error(err)
	}

	_, err = tx.Get(ctx, testKey1)
	assert.ErrorIs(t, err, ds.ErrNotFound)
}

func TestTxnGetSizeOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, false)
	if err != nil {
		t.Error(err)
	}

	resp, err := tx.GetSize(ctx, testKey1)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(testValue1), resp)
}

func TestTxnGetSizeOfterPutOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, false)
	if err != nil {
		t.Error(err)
	}

	err = tx.Put(ctx, testKey3, testValue3)
	if err != nil {
		t.Error(err)
	}

	resp, err := tx.GetSize(ctx, testKey3)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(testValue1), resp)
}

func TestTxnGetSizeOperationAfterDelete(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, false)
	if err != nil {
		t.Error(err)
	}

	err = tx.Delete(ctx, testKey1)
	if err != nil {
		t.Error(err)
	}

	_, err = tx.GetSize(ctx, testKey1)
	assert.ErrorIs(t, err, ds.ErrNotFound)
}

func TestTxnGetSizeOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, false)
	if err != nil {
		t.Error(err)
	}

	_, err = tx.GetSize(ctx, testKey3)
	assert.ErrorIs(t, err, ds.ErrNotFound)
}

func TestTxnHasOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, false)
	if err != nil {
		t.Error(err)
	}

	resp, err := tx.Has(ctx, testKey1)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, true, resp)
}

func TestTxnHasOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, false)
	if err != nil {
		t.Error(err)
	}

	resp, err := tx.Has(ctx, testKey3)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, false, resp)
}

func TestTxnHasOfterPutOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, false)
	if err != nil {
		t.Error(err)
	}

	err = tx.Put(ctx, testKey3, testValue3)
	if err != nil {
		t.Error(err)
	}

	resp, err := tx.Has(ctx, testKey3)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, true, resp)
}

func TestTxnHasOperationAfterDelete(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, false)
	if err != nil {
		t.Error(err)
	}

	err = tx.Delete(ctx, testKey1)
	if err != nil {
		t.Error(err)
	}

	resp, err := tx.Has(ctx, testKey1)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, false, resp)
}

func TestTxnPutAndCommitOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, false)
	if err != nil {
		t.Error(err)
	}

	err = tx.Put(ctx, testKey3, testValue3)
	if err != nil {
		t.Error(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		t.Error(err)
	}

	resp, err := s.Has(ctx, testKey3)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, true, resp)
}

func TestTxnPutAndCommitOperationReadOnly(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, true)
	if err != nil {
		t.Error(err)
	}

	err = tx.Put(ctx, testKey3, testValue3)
	assert.ErrorIs(t, err, ErrReadOnlyTxn)

	err = tx.Commit(ctx)
	assert.ErrorIs(t, err, ErrReadOnlyTxn)
}

func TestTxnPutOperationReadOnly(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, true)
	if err != nil {
		t.Error(err)
	}

	err = tx.Put(ctx, testKey3, testValue3)
	assert.ErrorIs(t, err, ErrReadOnlyTxn)
}

func TestTxnQueryOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, false)
	if err != nil {
		t.Error(err)
	}

	err = s.Put(ctx, testKey4, testValue4)
	if err != nil {
		t.Error(err)
	}

	err = tx.Put(ctx, testKey3, testValue3)
	if err != nil {
		t.Error(err)
	}

	err = tx.Put(ctx, testKey2, testValue3)
	if err != nil {
		t.Error(err)
	}

	err = tx.Delete(ctx, testKey1)
	if err != nil {
		t.Error(err)
	}

	err = tx.Delete(ctx, testKey5)
	if err != nil {
		t.Error(err)
	}

	results, err := tx.Query(ctx, dsq.Query{
		Limit:  1,
		Offset: 1,
	})
	if err != nil {
		t.Error(err)
	}

	result, _ := results.NextSync()

	assert.Equal(t, testKey3.String(), result.Entry.Key)
	assert.Equal(t, testValue3, result.Entry.Value)
}

func TestTxnQueryWithOnlyOneOperation(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx, err := s.NewTransaction(ctx, false)
	if err != nil {
		t.Error(err)
	}

	err = s.Put(ctx, testKey4, testValue4)
	if err != nil {
		t.Error(err)
	}

	results, err := tx.Query(ctx, dsq.Query{})
	if err != nil {
		t.Error(err)
	}

	_, _ = results.NextSync()
	_, _ = results.NextSync()
	result, _ := results.NextSync()

	assert.Equal(t, testKey4.String(), result.Entry.Key)
	assert.Equal(t, testValue4, result.Entry.Value)
}

func TestTxnDiscardOperationNotFound(t *testing.T) {
	ctx := context.Background()
	s := newLoadedDatastore()
	tx := &basicTxn{
		ops:      btree.NewMap[string, op](2),
		ds:       s,
		readOnly: false,
	}

	err := tx.Put(ctx, testKey3, testValue3)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 1, tx.ops.Len())

	tx.Discard(ctx)
	assert.Equal(t, 0, tx.ops.Len())
}
