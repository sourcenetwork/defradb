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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// filtered fetcher is responsible for the filtering documents based on the provided
// conditions.
type filtered struct {
	filter  *mapper.Filter
	mapping *core.DocumentMapping

	fetcher fetcher
}

var _ fetcher = (*filtered)(nil)

func newFilteredFetcher(
	filter *mapper.Filter,
	mapping *core.DocumentMapping,
	fetcher fetcher,
) *filtered {
	return &filtered{
		filter:  filter,
		mapping: mapping,
		fetcher: fetcher,
	}
}

func (f *filtered) NextDoc() (immutable.Option[string], error) {
	return f.fetcher.NextDoc()
}

func (f *filtered) GetFields() (immutable.Option[EncodedDocument], error) {
	doc, err := f.fetcher.GetFields()
	if err != nil {
		return immutable.None[EncodedDocument](), err
	}

	if !doc.HasValue() {
		return immutable.None[EncodedDocument](), nil
	}

	decodedDoc, err := DecodeToDoc(doc.Value(), f.mapping, false)
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

func (f *filtered) Close() error {
	return f.fetcher.Close()
}
