// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package fulltextindex adapts DefraDB transactions and keyspaces to full-text algorithms. It is
// the only database-facing package that knows the full-text storage layout.
package fulltextindex

import (
	"context"
	"errors"

	"github.com/sourcenetwork/defradb/client"
)

var (
	ErrUnsupportedAlgorithm = errors.New("unsupported full-text index algorithm")
	ErrMissingPosting       = errors.New("full-text index posting is missing")
)

// Index is a transaction-bound full-text index.
type Index interface {
	Insert(docShortID uint64, text string) error
	// Delete returns false when this document has no length/posting state. allowMissingPostings is
	// used only while a generic backfill is in progress, where a live delete may race ahead of it.
	Delete(docShortID uint64, text string, allowMissingPostings bool) (bool, error)
	// Search returns every matching document in descending relevance order. The concrete algorithm
	// owns its query analysis, storage reads, and scoring semantics.
	Search(query string) (SearchResult, error)
}

// Open binds one persisted full-text description and epoch to the transaction carried by ctx.
func Open(
	ctx context.Context,
	collectionShortID, indexID, epoch uint32,
	desc client.FullTextIndexDescription,
) (Index, error) {
	switch desc.Algorithm {
	case client.FullTextAlgorithmBM25:
		if desc.BM25 == nil {
			return nil, ErrUnsupportedAlgorithm
		}
		return &bm25Index{
			ctx:               ctx,
			collectionShortID: collectionShortID,
			indexID:           indexID,
			epoch:             epoch,
			params:            *desc.BM25,
		}, nil
	default:
		return nil, ErrUnsupportedAlgorithm
	}
}
