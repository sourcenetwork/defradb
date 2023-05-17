package mocks

import (
	"testing"

	ds "github.com/ipfs/go-datastore"
	query "github.com/ipfs/go-datastore/query"
	mock "github.com/stretchr/testify/mock"
)

type MultiStoreTxn struct {
	*Txn
	MockRootstore   *DSReaderWriter
	MockDatastore   *DSReaderWriter
	MockHeadstore   *DSReaderWriter
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

func prepareRootStore(t *testing.T) *DSReaderWriter {
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

	result := &MultiStoreTxn{
		Txn:             txn,
		MockRootstore:   prepareRootStore(t),
		MockDatastore:   prepareDataStore(t),
		MockHeadstore:   prepareHeadStore(t),
		MockDAGstore:    prepareDAGStore(t),
		MockSystemstore: prepareSystemStore(t),
	}

	txn.EXPECT().Rootstore().Return(result.MockRootstore).Maybe()
	txn.EXPECT().Datastore().Return(result.MockDatastore).Maybe()
	txn.EXPECT().Headstore().Return(result.MockHeadstore).Maybe()
	txn.EXPECT().DAGstore().Return(result.MockDAGstore).Maybe()
	txn.EXPECT().Systemstore().Return(result.MockSystemstore).Maybe()

	return result
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
