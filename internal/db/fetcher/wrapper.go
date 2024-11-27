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

// wrapper is a fetcher type that bridges between the existing [Fetcher] interface
// and the newer [fetcher] interface.
type wrapper struct {
	fetcher  fetcher
	execInfo *ExecInfo

	// The below properties are only held in state in order to temporarily adhear to the [Fetcher]
	// interface.  They can be remove from state once the [Fetcher] interface is cleaned up.
	identity    immutable.Option[acpIdentity.Identity]
	txn         datastore.Txn
	acp         immutable.Option[acp.ACP]
	col         client.Collection
	fields      []client.FieldDefinition
	filter      *mapper.Filter
	docMapper   *core.DocumentMapping
	showDeleted bool
}

var _ Fetcher = (*wrapper)(nil)

func NewDocumentFetcher() Fetcher {
	return &wrapper{}
}

func (f *wrapper) Init(
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	txn datastore.Txn,
	acp immutable.Option[acp.ACP],
	col client.Collection,
	fields []client.FieldDefinition,
	filter *mapper.Filter,
	docMapper *core.DocumentMapping,
	showDeleted bool,
) error {
	f.identity = identity
	f.txn = txn
	f.acp = acp
	f.col = col
	f.fields = fields
	f.filter = filter
	f.docMapper = docMapper
	f.showDeleted = showDeleted

	return nil
}

func (f *wrapper) Start(ctx context.Context, prefixes ...keys.Walkable) error {
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
		parsedfilterFields, err := parser.ParseFilterFieldsForDescription(conditions, f.col.Definition())
		if err != nil {
			return err
		}

		existingFields := make(map[client.FieldID]struct{}, len(f.fields))
		for _, field := range f.fields {
			existingFields[field.ID] = struct{}{}
		}

		for _, field := range parsedfilterFields {
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

	var execInfo ExecInfo
	f.execInfo = &execInfo

	var fetcher fetcher
	fetcher, err = newPrefixFetcher(ctx, f.txn, dsPrefixes, f.col, fieldsByID, client.Active, &execInfo)
	if err != nil {
		return nil
	}

	if f.showDeleted {
		deletedFetcher, err := newPrefixFetcher(ctx, f.txn, dsPrefixes, f.col, fieldsByID, client.Deleted, &execInfo)
		if err != nil {
			return nil
		}

		fetcher = newDeletedFetcher(fetcher, deletedFetcher)
	}

	if f.acp.HasValue() {
		fetcher = newPermissionedFetcher(ctx, f.identity, f.acp.Value(), f.col, fetcher)
	}

	if f.filter != nil {
		fetcher = newFilteredFetcher(f.filter, f.docMapper, fetcher)
	}

	f.fetcher = fetcher
	return nil
}

func (f *wrapper) FetchNext(ctx context.Context) (EncodedDocument, ExecInfo, error) {
	docID, err := f.fetcher.NextDoc()
	if err != nil {
		return nil, ExecInfo{}, err
	}

	if !docID.HasValue() {
		execInfo := *f.execInfo
		f.execInfo.Reset()

		return nil, execInfo, nil
	}

	doc, err := f.fetcher.GetFields()
	if err != nil {
		return nil, ExecInfo{}, err
	}

	if !doc.HasValue() {
		return f.FetchNext(ctx)
	}

	execInfo := *f.execInfo
	f.execInfo.Reset()

	return doc.Value(), execInfo, nil
}

func (f *wrapper) Close() error {
	if f.fetcher != nil {
		return f.fetcher.Close()
	}
	return nil
}
