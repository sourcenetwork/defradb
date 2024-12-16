// Copyright 2022 Democratized Data Foundation
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
)

// ExecInfo contains statistics about the fetcher execution.
type ExecInfo struct {
	// Number of documents fetched.
	DocsFetched uint64
	// Number of fields fetched.
	FieldsFetched uint64
	// Number of indexes fetched.
	IndexesFetched uint64
}

// Add adds the other ExecInfo to the current ExecInfo.
func (s *ExecInfo) Add(other ExecInfo) {
	s.DocsFetched += other.DocsFetched
	s.FieldsFetched += other.FieldsFetched
	s.IndexesFetched += other.IndexesFetched
}

// Reset resets the ExecInfo.
func (s *ExecInfo) Reset() {
	s.DocsFetched = 0
	s.FieldsFetched = 0
	s.IndexesFetched = 0
}

// Fetcher is the interface for collecting documents from the underlying data store.
// It handles all the key/value scanning, aggregation, and document encoding.
type Fetcher interface {
	Init(
		ctx context.Context,
		identity immutable.Option[acpIdentity.Identity],
		txn datastore.Txn,
		acp immutable.Option[acp.ACP],
		col client.Collection,
		fields []client.FieldDefinition,
		filter *mapper.Filter,
		docmapper *core.DocumentMapping,
		showDeleted bool,
	) error
	Start(ctx context.Context, prefixes ...keys.Walkable) error
	FetchNext(ctx context.Context) (EncodedDocument, ExecInfo, error)
	Close() error
}

// fetcher fetches documents from the store, performing low-level filtering
// when appropriate (e.g. ACP).
type fetcher interface {
	// NextDoc progresses the internal iterator(s) to the next document, yielding
	// its docID if found.
	//
	// If None is returned, iteration is complete and there are no more documents left
	// to fetch.
	NextDoc() (immutable.Option[string], error)

	// GetFields returns the EncodedDocument for the last docID yielded from [NextDoc()].
	//
	// If the field values for that document do not pass all filters, None will be returned -
	// this does not indicate that iteration has been completed, new documents may still be yielded
	// by [NextDoc()].
	GetFields() (immutable.Option[EncodedDocument], error)

	// Close disposes of all resources used by this fetcher and its children.
	Close() error
}
