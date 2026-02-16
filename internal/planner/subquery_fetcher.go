// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

import (
	"context"
	"errors"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/dac"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/datastore"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/lens"
	"github.com/sourcenetwork/defradb/internal/planner/filter"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
	lensStore "github.com/sourcenetwork/lens/host-go/store"
)

// subQueryFetcher executes sub-queries independently without mutating scan nodes.
// It creates its own fetcher instance for each query, avoiding shared state issues
// that can arise from the save-restore pattern.
type subQueryFetcher struct {
	ctx         context.Context
	identity    immutable.Option[acpIdentity.Identity]
	nodeACP     acpDB.NACInfo
	documentACP immutable.Option[dac.DocumentACP]
	col         client.Collection
	docMapping  *core.DocumentMapping
	lensStore   lensStore.Store

	// fields to fetch for each document
	fields []client.CollectionFieldDescription

	// execInfo accumulates fetch stats across all fetches
	execInfo *fetcher.ExecInfo
}

// newSubQueryFetcher creates a fetcher for sub-query execution.
// The execInfo parameter allows stats to be accumulated across multiple fetches.
func newSubQueryFetcher(
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	nodeACP acpDB.NACInfo,
	documentACP immutable.Option[dac.DocumentACP],
	col client.Collection,
	docMapping *core.DocumentMapping,
	lensStore lensStore.Store,
	fields []client.CollectionFieldDescription,
	execInfo *fetcher.ExecInfo,
) *subQueryFetcher {
	return &subQueryFetcher{
		ctx:         ctx,
		identity:    identity,
		nodeACP:     nodeACP,
		documentACP: documentACP,
		col:         col,
		docMapping:  docMapping,
		lensStore:   lensStore,
		fields:      fields,
		execInfo:    execInfo,
	}
}

// createFetcher creates a new fetcher instance wrapped with lens support.
func (f *subQueryFetcher) createFetcher() fetcher.Fetcher {
	baseFetcher := fetcher.NewDocumentFetcher()
	return lens.NewFetcher(baseFetcher, f.lensStore)
}

// fetchOrphans fetches documents where the relation ID field is NULL.
// These are "orphan" documents that don't have a related document on the other side.
// This is useful for queries that order by a relation field - orphans have NULL values
// and should be included in the results.
func (f *subQueryFetcher) fetchOrphans(
	relIDFieldName string,
	relIDFieldMapIndex int,
	filter *mapper.Filter,
) (docs []core.Doc, err error) {
	txn := datastore.CtxMustGetTxn(f.ctx)

	shortID, err := id.GetShortCollectionID(f.ctx, f.col.Version().CollectionID)
	if err != nil {
		return nil, err
	}

	filterWithNull := addNullFilterOnField(filter, relIDFieldMapIndex)

	result := selectIndex(selectIndexOptions{
		collection:          f.col,
		filter:              filterWithNull,
		relationIDFieldName: relIDFieldName,
		docMapping:          f.docMapping,
	})

	fetch := f.createFetcher()
	defer func() {
		err = errors.Join(err, fetch.Close())
	}()

	err = fetch.Init(
		f.ctx,
		f.identity,
		txn,
		f.nodeACP,
		f.documentACP,
		result.index,
		f.col,
		f.fields,
		filterWithNull,
		nil, // no ordering for orphans - they all have NULL values
		f.docMapping,
		false, // showDeleted
	)
	if err != nil {
		return nil, err
	}

	prefix := keys.DataStoreKey{CollectionShortID: shortID}
	err = fetch.Start(f.ctx, prefix)
	if err != nil {
		return nil, err
	}

	return f.collectAllDocs(fetch, shortID)
}

// fetchAllExcluding fetches all documents from the collection, excluding those with IDs in excludeIDs.
// This is used to find "orphan" documents when the collection doesn't have a foreign key field
// (i.e., the collection is on the secondary side of a relation).
func (f *subQueryFetcher) fetchAllExcluding(
	filter *mapper.Filter,
	excludeIDs []string,
) (docs []core.Doc, err error) {
	txn := datastore.CtxMustGetTxn(f.ctx)

	shortID, err := id.GetShortCollectionID(f.ctx, f.col.Version().CollectionID)
	if err != nil {
		return nil, err
	}

	result := selectIndex(selectIndexOptions{
		collection: f.col,
		filter:     filter,
		docMapping: f.docMapping,
	})

	fetch := f.createFetcher()
	defer func() {
		err = errors.Join(err, fetch.Close())
	}()

	err = fetch.Init(
		f.ctx,
		f.identity,
		txn,
		f.nodeACP,
		f.documentACP,
		result.index,
		f.col,
		f.fields,
		filter,
		nil, // no ordering
		f.docMapping,
		false, // showDeleted
	)
	if err != nil {
		return nil, err
	}

	prefix := keys.DataStoreKey{CollectionShortID: shortID}
	err = fetch.Start(f.ctx, prefix)
	if err != nil {
		return nil, err
	}

	return f.collectAllDocsExcluding(fetch, shortID, excludeIDs)
}

// fetchOrphansByParentConstraint fetches orphan documents for a specific parent.
// It first fetches all primary docs that reference the given secondary doc (parent constraint),
// then filters to only return those that are not already in existingIDs and have a NULL relation field
// (i.e., orphans without the ordering relation).
func (f *subQueryFetcher) fetchOrphansByParentConstraint(
	baseFilter *mapper.Filter,
	relIDFieldMapIndex int,
	relFieldIndex int,
	parentDocID string,
	relationIDFieldName string,
	existingIDs map[string]struct{},
) (docs []core.Doc, err error) {
	txn := datastore.CtxMustGetTxn(f.ctx)

	shortID, err := id.GetShortCollectionID(f.ctx, f.col.Version().CollectionID)
	if err != nil {
		return nil, err
	}

	parentFilter := addFilterOnField(baseFilter, relIDFieldMapIndex, parentDocID)

	result := selectIndex(selectIndexOptions{
		collection:          f.col,
		filter:              parentFilter,
		relationIDFieldName: relationIDFieldName,
		docMapping:          f.docMapping,
	})

	fetch := f.createFetcher()
	defer func() {
		err = errors.Join(err, fetch.Close())
	}()

	err = fetch.Init(
		f.ctx,
		f.identity,
		txn,
		f.nodeACP,
		f.documentACP,
		result.index,
		f.col,
		f.fields,
		parentFilter,
		nil, // no ordering for orphan scan
		f.docMapping,
		false, // showDeleted
	)
	if err != nil {
		return nil, err
	}

	prefix := keys.DataStoreKey{CollectionShortID: shortID}
	err = fetch.Start(f.ctx, prefix)
	if err != nil {
		return nil, err
	}

	allDocs, err := f.collectAllDocs(fetch, shortID)
	if err != nil {
		return nil, err
	}

	orphanDocs := make([]core.Doc, 0)
	for _, doc := range allDocs {
		if _, exists := existingIDs[doc.GetID()]; exists {
			continue
		}
		if doc.Fields[relFieldIndex] == nil {
			orphanDocs = append(orphanDocs, doc)
		}
	}

	return orphanDocs, nil
}

// addFilterOnField adds a filter condition that checks if the field equals the given value.
func addFilterOnField(f *mapper.Filter, propIndex int, value any) *mapper.Filter {
	if f == nil {
		f = mapper.NewFilter()
	}

	propertyIndex := &mapper.PropertyIndex{Index: propIndex}
	filterConditions := map[connor.FilterKey]any{
		propertyIndex: map[connor.FilterKey]any{
			mapper.FilterEqOp: value,
		},
	}

	filter.RemoveField(f, mapper.Field{Index: propIndex})
	f.Conditions = filter.MergeConditions(f.Conditions, filterConditions)
	return f
}

// addNullFilterOnField adds a filter condition that checks if the field is NULL.
func addNullFilterOnField(f *mapper.Filter, propIndex int) *mapper.Filter {
	return addFilterOnField(f, propIndex, nil)
}

// collectAllDocs fetches all documents from the fetcher.
func (f *subQueryFetcher) collectAllDocs(fetch fetcher.Fetcher, shortID uint32) ([]core.Doc, error) {
	return f.collectAllDocsExcluding(fetch, shortID, nil)
}

// collectAllDocsExcluding fetches all documents from the fetcher, excluding those with IDs in the excludeIDs list.
func (f *subQueryFetcher) collectAllDocsExcluding(fetch fetcher.Fetcher, shortID uint32, excludeIDs []string) ([]core.Doc, error) {
	var docs []core.Doc

	excludeSet := make(map[string]struct{}, len(excludeIDs))
	for _, id := range excludeIDs {
		excludeSet[id] = struct{}{}
	}

	for {
		encDoc, fetchExecInfo, err := fetch.FetchNext(f.ctx)
		if err != nil {
			return nil, err
		}

		if f.execInfo != nil {
			f.execInfo.Add(fetchExecInfo)
		}

		if encDoc == nil {
			break
		}

		doc, err := fetcher.DecodeToDoc(f.ctx, shortID, encDoc, f.docMapping, false)
		if err != nil {
			return nil, err
		}

		if _, excluded := excludeSet[doc.GetID()]; excluded {
			continue
		}

		docs = append(docs, doc)
	}

	return docs, nil
}
