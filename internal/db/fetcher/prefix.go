// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fetcher

import (
	"context"
	"slices"
	"strings"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// prefixFetcher is a fetcher type responsible for iterating through multiple prefixes.
//
// It manages the document fetcher instances that will do the actual scanning.
type prefixFetcher struct {
	// The prefixes that this prefix fetcher must fetch from.
	prefixes []keys.DataStoreKey

	// The index of the current prefix being fetched.
	currentPrefix int
	// The child document fetcher, specific to the current prefix.
	fetcher *documentFetcher

	// The below properties are only held here in order to pass them on to the next
	// child fetcher instance.
	ctx        context.Context
	txn        datastore.Txn
	fieldsByID map[uint32]client.FieldDefinition
	status     client.DocumentStatus
	execInfo   *ExecInfo
}

var _ fetcher = (*prefixFetcher)(nil)

func newPrefixFetcher(
	ctx context.Context,
	txn datastore.Txn,
	prefixes []keys.DataStoreKey,
	col client.Collection,
	fieldsByID map[uint32]client.FieldDefinition,
	status client.DocumentStatus,
	execInfo *ExecInfo,
) (*prefixFetcher, error) {
	if len(prefixes) == 0 {
		shortID, err := id.GetShortCollectionID(ctx, col.Version().CollectionID)
		if err != nil {
			return nil, err
		}

		// If no prefixes are provided, scan the entire collection.
		prefixes = append(prefixes, keys.DataStoreKey{
			CollectionShortID: shortID,
		})
	} else {
		uniquePrefixes := make(map[keys.DataStoreKey]struct{}, len(prefixes))
		for _, prefix := range prefixes {
			// Deduplicate the prefixes to make sure that any given document is only yielded
			// once.
			uniquePrefixes[prefix] = struct{}{}
		}

		prefixes = make([]keys.DataStoreKey, 0, len(uniquePrefixes))
		for prefix := range uniquePrefixes {
			prefixes = append(prefixes, prefix)
		}

		// Sort the prefixes, so that documents are returned in the order they would be if the
		// whole store was scanned.
		slices.SortFunc(prefixes, func(a, b keys.DataStoreKey) int {
			return strings.Compare(a.ToString(), b.ToString())
		})
	}

	fetcher, err := newDocumentFetcher(ctx, txn, fieldsByID, prefixes[0], status, execInfo)
	if err != nil {
		return nil, err
	}

	return &prefixFetcher{
		txn:        txn,
		prefixes:   prefixes,
		ctx:        ctx,
		fieldsByID: fieldsByID,
		status:     status,
		fetcher:    fetcher,
		execInfo:   execInfo,
	}, nil
}

func (f *prefixFetcher) NextDoc() (immutable.Option[string], error) {
	docID, err := f.fetcher.NextDoc()
	if err != nil {
		return immutable.None[string](), err
	}

	if !docID.HasValue() {
		f.currentPrefix++
		if f.fetcher != nil {
			err := f.fetcher.Close()
			if err != nil {
				return immutable.None[string](), err
			}
		}

		if len(f.prefixes) > f.currentPrefix {
			prefix := f.prefixes[f.currentPrefix]

			fetcher, err := newDocumentFetcher(f.ctx, f.txn, f.fieldsByID, prefix, f.status, f.execInfo)
			if err != nil {
				return immutable.None[string](), err
			}
			f.fetcher = fetcher

			return f.NextDoc()
		}
	}

	return docID, nil
}

func (f *prefixFetcher) GetFields() (immutable.Option[EncodedDocument], error) {
	return f.fetcher.GetFields()
}

func (f *prefixFetcher) Close() error {
	return f.fetcher.Close()
}
