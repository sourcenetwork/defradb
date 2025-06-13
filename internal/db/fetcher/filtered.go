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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// filteredFetcher fetcher is responsible for the filtering documents based on the provided
// conditions.
type filteredFetcher struct {
	ctx               context.Context
	collectionShortID uint32
	filter            *mapper.Filter
	mapping           *core.DocumentMapping

	fetcher fetcher
}

var _ fetcher = (*filteredFetcher)(nil)

func newFilteredFetcher(
	ctx context.Context,
	collectionShortID uint32,
	filter *mapper.Filter,
	mapping *core.DocumentMapping,
	fetcher fetcher,
) *filteredFetcher {
	return &filteredFetcher{
		ctx:               ctx,
		collectionShortID: collectionShortID,
		filter:            filter,
		mapping:           mapping,
		fetcher:           fetcher,
	}
}

func (f *filteredFetcher) NextDoc() (immutable.Option[string], error) {
	return f.fetcher.NextDoc()
}

func (f *filteredFetcher) GetFields() (immutable.Option[EncodedDocument], error) {
	doc, err := f.fetcher.GetFields()
	if err != nil {
		return immutable.None[EncodedDocument](), err
	}

	if !doc.HasValue() {
		return immutable.None[EncodedDocument](), nil
	}

	decodedDoc, err := DecodeToDoc(f.ctx, f.collectionShortID, doc.Value(), f.mapping, false)
	if err != nil {
		return immutable.None[EncodedDocument](), err
	}

	passedFilter, err := mapper.RunFilter(decodedDoc, f.filter)
	if err != nil {
		return immutable.None[EncodedDocument](), err
	}

	if !passedFilter {
		return immutable.None[EncodedDocument](), nil
	}

	return doc, nil
}

func (f *filteredFetcher) Close() error {
	return f.fetcher.Close()
}
