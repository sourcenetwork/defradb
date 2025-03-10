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

	"github.com/sourcenetwork/defradb/acp"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
	"github.com/sourcenetwork/defradb/internal/request/graphql/parser"
)

// wrappingFetcher is a fetcher type that bridges between the existing [Fetcher] interface
// and the newer [fetcher] interface.
type wrappingFetcher struct {
	acp       immutable.Option[acp.ACP]
	fetcher   fetcher
	txn       datastore.Txn
	col       client.Collection
	filter    *mapper.Filter
	docMapper *core.DocumentMapping

	// The below properties are only held in state in order to temporarily adhere to the [Fetcher]
	// interface.  They can be remove from state once the [Fetcher] interface is cleaned up.
	identity immutable.Option[acpIdentity.Identity]
	fields   []client.FieldDefinition
	index    immutable.Option[client.IndexDescription]
	execInfo ExecInfo

	showDeleted bool
}

var _ Fetcher = (*wrappingFetcher)(nil)

func NewDocumentFetcher() Fetcher {
	return &wrappingFetcher{}
}

func (f *wrappingFetcher) Init(
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	txn datastore.Txn,
	acp immutable.Option[acp.ACP],
	index immutable.Option[client.IndexDescription],
	col client.Collection,
	fields []client.FieldDefinition,
	filter *mapper.Filter,
	docMapper *core.DocumentMapping,
	showDeleted bool,
) error {
	f.identity = identity
	f.txn = txn
	f.acp = acp
	f.index = index
	f.col = col
	f.fields = fields
	f.filter = filter
	f.docMapper = docMapper
	f.showDeleted = showDeleted

	return nil
}

func (f *wrappingFetcher) Start(ctx context.Context, prefixes ...keys.Walkable) error {
	err := f.Close()
	if err != nil {
		return err
	}

	dsPrefixes := make([]keys.DataStoreKey, 0, len(prefixes))
	for _, prefix := range prefixes {
		dsPrefix, ok := prefix.(keys.DataStoreKey)
		if !ok {
			continue
		}

		dsPrefixes = append(dsPrefixes, dsPrefix)
	}

	if f.filter != nil && len(f.fields) > 0 {
		conditions := f.filter.ToMap(f.docMapper)
		parsedFilterFields, err := parser.ParseFilterFieldsForDescription(conditions, f.col.Definition())
		if err != nil {
			return err
		}

		existingFields := make(map[client.FieldID]struct{}, len(f.fields))
		for _, field := range f.fields {
			existingFields[field.ID] = struct{}{}
		}

		for _, field := range parsedFilterFields {
			if _, ok := existingFields[field.ID]; !ok {
				f.fields = append(f.fields, field)
			}
			existingFields[field.ID] = struct{}{}
		}
	}

	if len(f.fields) == 0 {
		f.fields = f.col.Definition().GetFields()
	}

	fieldsByID := make(map[uint32]client.FieldDefinition, len(f.fields))
	for _, field := range f.fields {
		fieldsByID[uint32(field.ID)] = field
	}

	f.execInfo.Reset()

	var top fetcher
	if f.index.HasValue() {
		indexFetcher, err := newIndexFetcher(ctx, f.txn, fieldsByID, f.index.Value(), f.filter, f.col,
			f.docMapper, &f.execInfo)
		if err != nil {
			return err
		}
		if indexFetcher != nil {
			top = indexFetcher
		}
	}

	// the index fetcher might not have been created if there is no efficient way to use fetch indexes
	// with given filter conditions. In this case we fall back to the prefix fetcher
	if top == nil {
		top, err = newPrefixFetcher(ctx, f.txn, dsPrefixes, f.col, fieldsByID, client.Active, &f.execInfo)
		if err != nil {
			return err
		}
	}

	if f.showDeleted {
		deletedFetcher, err := newPrefixFetcher(ctx, f.txn, dsPrefixes, f.col, fieldsByID, client.Deleted, &f.execInfo)
		if err != nil {
			return err
		}

		top = newMultiFetcher(top, deletedFetcher)
	}

	if f.acp.HasValue() {
		top = newPermissionedFetcher(ctx, f.identity, f.acp.Value(), f.col, top)
	}

	if f.filter != nil {
		top = newFilteredFetcher(f.filter, f.docMapper, top)
	}

	f.fetcher = top
	return nil
}

func (f *wrappingFetcher) FetchNext(ctx context.Context) (EncodedDocument, ExecInfo, error) {
	f.execInfo.Reset()

	for {
		docID, err := f.fetcher.NextDoc()
		if err != nil {
			return nil, ExecInfo{}, err
		}

		if !docID.HasValue() {
			return nil, f.execInfo, nil
		}

		doc, err := f.fetcher.GetFields()
		if err != nil {
			return nil, ExecInfo{}, err
		}

		if !doc.HasValue() {
			continue
		}

		return doc.Value(), f.execInfo, nil
	}
}

func (f *wrappingFetcher) Close() error {
	if f.fetcher != nil {
		return f.fetcher.Close()
	}
	return nil
}
