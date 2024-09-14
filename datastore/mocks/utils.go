// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package mocks

import (
	"testing"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/stretchr/testify/mock"
)

type MultiStoreTxn struct {
	*Txn
	t               *testing.T
	MockRootstore   *DSReaderWriter
	MockDatastore   *DSReaderWriter
	MockHeadstore   *DSReaderWriter
	MockEncstore    *Blockstore
	MockDAGstore    *DAGStore
	MockSystemstore *DSReaderWriter
}

func prepareDataStore(t *testing.T) *DSReaderWriter {
	dataStore := NewDSReaderWriter(t)
	dataStore.EXPECT().Get(mock.Anything, mock.Anything).Return([]byte{}, ds.ErrNotFound).Maybe()
	dataStore.EXPECT().Put(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	dataStore.EXPECT().Has(mock.Anything, mock.Anything).Return(true, nil).Maybe()
	return dataStore
}

func prepareEncStore(t *testing.T) *Blockstore {
	encStore := NewBlockstore(t)
	encStore.EXPECT().Get(mock.Anything, mock.Anything).Return(nil, ds.ErrNotFound).Maybe()
	encStore.EXPECT().Put(mock.Anything, mock.Anything).Return(nil).Maybe()
	encStore.EXPECT().Has(mock.Anything, mock.Anything).Return(true, nil).Maybe()
	return encStore
}

func prepareRootstore(t *testing.T) *DSReaderWriter {
	return NewDSReaderWriter(t)
}

func prepareHeadStore(t *testing.T) *DSReaderWriter {
	headStore := NewDSReaderWriter(t)

	headStore.EXPECT().Query(mock.Anything, mock.Anything).
		Return(NewQueryResultsWithValues(t), nil).Maybe()

	headStore.EXPECT().Get(mock.Anything, mock.Anything).Return([]byte{}, ds.ErrNotFound).Maybe()
	headStore.EXPECT().Put(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	headStore.EXPECT().Has(mock.Anything, mock.Anything).Return(false, nil).Maybe()
	return headStore
}

func prepareSystemStore(t *testing.T) *DSReaderWriter {
	systemStore := NewDSReaderWriter(t)
	systemStore.EXPECT().Get(mock.Anything, mock.Anything).Return([]byte{}, nil).Maybe()
	return systemStore
}

func prepareDAGStore(t *testing.T) *DAGStore {
	dagStore := NewDAGStore(t)
	dagStore.EXPECT().Put(mock.Anything, mock.Anything).Return(nil).Maybe()
	dagStore.EXPECT().Has(mock.Anything, mock.Anything).Return(false, nil).Maybe()
	return dagStore
}

func NewTxnWithMultistore(t *testing.T) *MultiStoreTxn {
	txn := NewTxn(t)
	txn.EXPECT().OnSuccess(mock.Anything).Maybe()
	txn.EXPECT().OnSuccessAsync(mock.Anything).Maybe()

	result := &MultiStoreTxn{
		Txn:             txn,
		t:               t,
		MockRootstore:   prepareRootstore(t),
		MockDatastore:   prepareDataStore(t),
		MockEncstore:    prepareEncStore(t),
		MockHeadstore:   prepareHeadStore(t),
		MockDAGstore:    prepareDAGStore(t),
		MockSystemstore: prepareSystemStore(t),
	}

	txn.EXPECT().Rootstore().Return(result.MockRootstore).Maybe()
	txn.EXPECT().Datastore().Return(result.MockDatastore).Maybe()
	txn.EXPECT().Encstore().Return(result.MockEncstore).Maybe()
	txn.EXPECT().Headstore().Return(result.MockHeadstore).Maybe()
	txn.EXPECT().Blockstore().Return(result.MockDAGstore).Maybe()
	txn.EXPECT().Systemstore().Return(result.MockSystemstore).Maybe()

	return result
}

func (txn *MultiStoreTxn) ClearSystemStore() *MultiStoreTxn {
	txn.MockSystemstore = NewDSReaderWriter(txn.t)
	txn.EXPECT().Systemstore().Unset()
	txn.EXPECT().Systemstore().Return(txn.MockSystemstore).Maybe()
	return txn
}

func NewQueryResultsWithValues(t *testing.T, values ...[]byte) *Results {
	results := make([]query.Result, len(values))
	for i, value := range values {
		results[i] = query.Result{Entry: query.Entry{Value: value}}
	}
	return NewQueryResultsWithResults(t, results...)
}

func NewQueryResultsWithResults(t *testing.T, results ...query.Result) *Results {
	queryResults := NewResults(t)
	resultChan := make(chan query.Result, len(results))
	for _, result := range results {
		resultChan <- result
	}
	close(resultChan)
	queryResults.EXPECT().Next().Return(resultChan).Maybe()
	queryResults.EXPECT().Close().Return(nil).Maybe()
	return queryResults
}
