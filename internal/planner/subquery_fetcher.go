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
	"maps"

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

// fetchDocs runs a full fetch pipeline: selectIndex → Init → Start → collect.
// The filter and relationIDFieldName control index selection. excludeIDs filters out
// documents by ID during collection.
func (f *subQueryFetcher) fetchDocs(
	filter *mapper.Filter,
	relationIDFieldName string,
	excludeIDs []string,
) (docs []core.Doc, err error) {
	txn := datastore.CtxMustGetTxn(f.ctx)

	shortID, err := id.GetShortCollectionID(f.ctx, f.col.Version().CollectionID)
	if err != nil {
		return nil, err
	}

	result := selectIndex(selectIndexOptions{
		collection:          f.col,
		filter:              filter,
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
		filter,
		nil,
		f.docMapping,
		false,
	)
	if err != nil {
		return nil, err
	}

	prefix := keys.DataStoreKey{CollectionShortID: shortID}
	err = fetch.Start(f.ctx, prefix)
	if err != nil {
		return nil, err
	}

	return f.collectDocs(fetch, shortID, excludeIDs)
}

// fetchOrphans fetches documents where the relation ID field is NULL.
func (f *subQueryFetcher) fetchOrphans(
	relIDFieldName string,
	relIDFieldMapIndex int,
	filter *mapper.Filter,
) ([]core.Doc, error) {
	filterWithNull := addNullFilterOnField(filter, relIDFieldMapIndex)
	return f.fetchDocs(filterWithNull, relIDFieldName, nil)
}

// fetchAllExcluding fetches all documents from the collection, excluding those with IDs in excludeIDs.
func (f *subQueryFetcher) fetchAllExcluding(
	filter *mapper.Filter,
	excludeIDs []string,
) ([]core.Doc, error) {
	return f.fetchDocs(filter, "", excludeIDs)
}

// addFilterOnField returns a new filter with a condition that checks if the field equals the given value.
// It does not mutate the input filter.
func addFilterOnField(f *mapper.Filter, propIndex int, value any) *mapper.Filter {
	result := mapper.NewFilter()
	if f != nil {
		maps.Copy(result.Conditions, f.Conditions)
		result.ExternalConditions = make(map[string]any, len(f.ExternalConditions))
		maps.Copy(result.ExternalConditions, f.ExternalConditions)
	}

	propertyIndex := &mapper.PropertyIndex{Index: propIndex}
	filterConditions := map[connor.FilterKey]any{
		propertyIndex: map[connor.FilterKey]any{
			mapper.FilterEqOp: value,
		},
	}

	filter.RemoveField(result, mapper.Field{Index: propIndex})
	result.Conditions = filter.MergeConditions(result.Conditions, filterConditions)
	return result
}

// addNullFilterOnField adds a filter condition that checks if the field is NULL.
func addNullFilterOnField(f *mapper.Filter, propIndex int) *mapper.Filter {
	return addFilterOnField(f, propIndex, nil)
}

// collectDocs fetches all documents from the fetcher, optionally excluding those with IDs in excludeIDs.
func (f *subQueryFetcher) collectDocs(fetch fetcher.Fetcher, shortID uint32, excludeIDs []string) ([]core.Doc, error) {
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
