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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
)

type txnCollection struct {
	txn   *Txn
	inner client.Collection
}

var _ client.Collection = (*txnCollection)(nil)

func newTxnCollection(txn *Txn, col client.Collection) *txnCollection {
	return &txnCollection{
		txn:   txn,
		inner: col,
	}
}

func (col *txnCollection) Name() string {
	return col.inner.Name()
}

func (col *txnCollection) VersionID() string {
	return col.inner.VersionID()
}

func (col *txnCollection) CollectionID() string {
	return col.inner.CollectionID()
}

func (col *txnCollection) Version() client.CollectionVersion {
	return col.inner.Version()
}

func (col *txnCollection) AddDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.AddDocumentOptions],
) error {
	ctx, unlock := lockForTxn(ctx, col.txn)
	defer unlock()

	if col.txn.isClosed {
		return ErrTxnDiscarded
	}

	return col.inner.AddDocument(ctx, doc, opts...)
}

func (col *txnCollection) AddManyDocuments(
	ctx context.Context,
	docs []*client.Document,
	opts ...options.Enumerable[options.AddDocumentOptions],
) error {
	ctx, unlock := lockForTxn(ctx, col.txn)
	defer unlock()

	if col.txn.isClosed {
		return ErrTxnDiscarded
	}

	return col.inner.AddManyDocuments(ctx, docs, opts...)
}

func (col *txnCollection) UpdateDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.UpdateDocumentOptions],
) error {
	ctx, unlock := lockForTxn(ctx, col.txn)
	defer unlock()

	if col.txn.isClosed {
		return ErrTxnDiscarded
	}

	return col.inner.UpdateDocument(ctx, doc, opts...)
}

func (col *txnCollection) SaveDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.SaveDocumentOptions],
) error {
	ctx, unlock := lockForTxn(ctx, col.txn)
	defer unlock()

	if col.txn.isClosed {
		return ErrTxnDiscarded
	}

	return col.inner.SaveDocument(ctx, doc, opts...)
}

func (col *txnCollection) DeleteDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.DeleteDocumentOptions],
) (bool, error) {
	ctx, unlock := lockForTxn(ctx, col.txn)
	defer unlock()

	if col.txn.isClosed {
		return false, ErrTxnDiscarded
	}

	return col.inner.DeleteDocument(ctx, docID, opts...)
}

func (col *txnCollection) ExistsDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.ExistsDocumentOptions],
) (bool, error) {
	ctx, unlock := lockForTxn(ctx, col.txn)
	defer unlock()

	if col.txn.isClosed {
		return false, ErrTxnDiscarded
	}

	return col.inner.ExistsDocument(ctx, docID, opts...)
}

func (col *txnCollection) UpdateDocumentsWithFilter(
	ctx context.Context,
	filter any,
	updater string,
	opts ...options.Enumerable[options.UpdateDocumentsWithFilterOptions],
) (*client.UpdateResult, error) {
	ctx, unlock := lockForTxn(ctx, col.txn)
	defer unlock()

	if col.txn.isClosed {
		return nil, ErrTxnDiscarded
	}

	return col.inner.UpdateDocumentsWithFilter(ctx, filter, updater, opts...)
}

func (col *txnCollection) DeleteDocumentsWithFilter(
	ctx context.Context,
	filter any,
	opts ...options.Enumerable[options.DeleteDocumentsWithFilterOptions],
) (*client.DeleteResult, error) {
	ctx, unlock := lockForTxn(ctx, col.txn)
	defer unlock()

	if col.txn.isClosed {
		return nil, ErrTxnDiscarded
	}

	return col.inner.DeleteDocumentsWithFilter(ctx, filter, opts...)
}

func (col *txnCollection) GetDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.GetDocumentOptions],
) (*client.Document, error) {
	ctx, unlock := lockForTxn(ctx, col.txn)
	defer unlock()

	if col.txn.isClosed {
		return nil, ErrTxnDiscarded
	}

	return col.inner.GetDocument(ctx, docID, opts...)
}

func (col *txnCollection) NewIndex(
	ctx context.Context,
	req client.NewIndexRequest,
	opts ...options.Enumerable[options.NewCollectionIndexOptions],
) (client.IndexDescription, error) {
	ctx, unlock := lockForTxn(ctx, col.txn)
	defer unlock()

	if col.txn.isClosed {
		return client.IndexDescription{}, ErrTxnDiscarded
	}

	return col.inner.NewIndex(ctx, req, opts...)
}

func (col *txnCollection) DeleteIndex(
	ctx context.Context,
	indexName string,
	opts ...options.Enumerable[options.DeleteCollectionIndexOptions],
) error {
	ctx, unlock := lockForTxn(ctx, col.txn)
	defer unlock()

	if col.txn.isClosed {
		return ErrTxnDiscarded
	}

	return col.inner.DeleteIndex(ctx, indexName, opts...)
}

func (col *txnCollection) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListCollectionIndexesOptions],
) ([]client.IndexDescription, error) {
	ctx, unlock := lockForTxn(ctx, col.txn)
	defer unlock()

	if col.txn.isClosed {
		return nil, ErrTxnDiscarded
	}

	return col.inner.ListIndexes(ctx, opts...)
}

func (col *txnCollection) NewEncryptedIndex(
	ctx context.Context,
	desc client.EncryptedIndexDescription,
	opts ...options.Enumerable[options.NewEncryptedIndexOptions],
) (client.EncryptedIndexDescription, error) {
	ctx, unlock := lockForTxn(ctx, col.txn)
	defer unlock()

	if col.txn.isClosed {
		return client.EncryptedIndexDescription{}, ErrTxnDiscarded
	}

	return col.inner.NewEncryptedIndex(ctx, desc, opts...)
}

func (col *txnCollection) DeleteEncryptedIndex(
	ctx context.Context,
	fieldName string,
	opts ...options.Enumerable[options.DeleteEncryptedIndexOptions],
) error {
	ctx, unlock := lockForTxn(ctx, col.txn)
	defer unlock()

	if col.txn.isClosed {
		return ErrTxnDiscarded
	}

	return col.inner.DeleteEncryptedIndex(ctx, fieldName, opts...)
}

func (col *txnCollection) ListEncryptedIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListCollectionEncryptedIndexesOptions],
) ([]client.EncryptedIndexDescription, error) {
	ctx, unlock := lockForTxn(ctx, col.txn)
	defer unlock()

	if col.txn.isClosed {
		return nil, ErrTxnDiscarded
	}

	return col.inner.ListEncryptedIndexes(ctx, opts...)
}

func (col *txnCollection) Truncate(
	ctx context.Context,
	opts ...options.Enumerable[options.TruncateCollectionOptions],
) error {
	ctx, unlock := lockForTxn(ctx, col.txn)
	defer unlock()

	if col.txn.isClosed {
		return ErrTxnDiscarded
	}

	return col.inner.Truncate(ctx, opts...)
}
