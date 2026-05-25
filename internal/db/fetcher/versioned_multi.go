// Copyright 2026 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/acp/dac"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/datastore"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
	"github.com/sourcenetwork/immutable"
)

// MultiVersioned is a short term solution that allows multiple prefixes to be
// fetched via the VersionedFetcher(s).
//
// It and the VersionedFetcher need  a rework at some point, migrating them to the
// `fetcher` interface, and resolving a few other their inefficiencies (such as the
// one that this type represents).
type MultiVersioned struct {
	children     []*VersionedFetcher
	currentChild int

	ctx         context.Context
	identity    immutable.Option[acpIdentity.Identity]
	txn         datastore.Txn
	nodeACP     acpDB.NACInfo
	documentACP immutable.Option[dac.DocumentACP]
	index       immutable.Option[client.IndexDescription]
	col         client.Collection
	fields      []client.CollectionFieldDescription
	filter      *mapper.Filter
	ordering    []mapper.OrderCondition
	docmapper   *core.DocumentMapping
	showDeleted bool
}

var _ Fetcher = (*MultiVersioned)(nil)

func (f *MultiVersioned) Init(
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	txn datastore.Txn,
	nodeACP acpDB.NACInfo,
	documentACP immutable.Option[dac.DocumentACP],
	index immutable.Option[client.IndexDescription],
	col client.Collection,
	fields []client.CollectionFieldDescription,
	filter *mapper.Filter,
	ordering []mapper.OrderCondition,
	docmapper *core.DocumentMapping,
	showDeleted bool,
) error {
	f.ctx = ctx
	f.identity = identity
	f.txn = txn
	f.nodeACP = nodeACP
	f.documentACP = documentACP
	f.index = index
	f.col = col
	f.fields = fields
	f.filter = filter
	f.ordering = ordering
	f.docmapper = docmapper
	f.showDeleted = showDeleted

	return nil
}

func (f *MultiVersioned) Start(ctx context.Context, prefixes ...keys.Walkable) error {
	// Deduplicate the prefixes so that the child fetchers do not return duplicated results.
	// This is in keeping with the behaviour of other query params, such as docIDs.
	uniquePrefixStrings := map[string]struct{}{}
	uniquePrefixes := []keys.Walkable{}
	for _, prefix := range prefixes {
		// Do not use `prefix.ToString()` here, that function returns a pretified result and there is
		// no guarantee that it will not clash with other unique prefixes.
		stringPrefix := string(prefix.Bytes())

		if _, ok := uniquePrefixStrings[stringPrefix]; ok {
			continue
		}

		uniquePrefixStrings[stringPrefix] = struct{}{}
		uniquePrefixes = append(uniquePrefixes, prefix)
	}

	f.children = make([]*VersionedFetcher, len(uniquePrefixes))

	for i, prefix := range uniquePrefixes {
		child := &VersionedFetcher{}
		err := child.Init(
			f.ctx,
			f.identity,
			f.txn,
			f.nodeACP,
			f.documentACP,
			f.index,
			f.col,
			f.fields,
			f.filter,
			f.ordering,
			f.docmapper,
			f.showDeleted,
		)
		if err != nil {
			return err
		}

		err = child.Start(ctx, prefix)
		if err != nil {
			return err
		}

		f.children[i] = child
	}

	f.currentChild = 0

	return nil
}

func (f *MultiVersioned) FetchNext(ctx context.Context) (EncodedDocument, ExecInfo, error) {
	if f.currentChild >= len(f.children) {
		return nil, ExecInfo{}, nil
	}

	doc, execInfo, err := f.children[f.currentChild].FetchNext(ctx)
	if err != nil {
		return nil, ExecInfo{}, err
	}

	if doc == nil {
		f.currentChild++
		return f.FetchNext(ctx)
	}

	return doc, execInfo, nil
}

func (f *MultiVersioned) Close() error {
	errs := []error{}
	for _, child := range f.children {
		if child == nil {
			// If an error is thrown in `f.Start`, this child, and later children in the loop might be nil.
			// If this child is nil, the later ones will be too.  If a child is nil, calling `child.Close`
			// will panic.
			break
		}

		err := child.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
