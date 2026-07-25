// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"

	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
)

// indexKindConstructor builds the implementation of one index kind from the shared index base.
type indexKindConstructor func(base collectionBaseIndex) client.CollectionIndex

// indexKinds maps an index kind to its implementation. It is written only by registerIndexKind
// during package initialisation, so the set of kinds a build supports is fixed at compile time.
var indexKinds = map[string]indexKindConstructor{}

// registerIndexKind registers the constructor for an index kind.
func registerIndexKind(kind string, constructor indexKindConstructor) {
	indexKinds[kind] = constructor
}

func init() {
	// The empty kind is the ordered key index that predates index kinds, so descriptions written
	// before the Kind field existed resolve to it. Unique and non-unique are two implementations
	// of that one kind, chosen by the description.
	registerIndexKind("", func(base collectionBaseIndex) client.CollectionIndex {
		if base.desc.Unique {
			return &collectionUniqueIndex{collectionBaseIndex: base}
		}
		return &collectionSimpleIndex{collectionBaseIndex: base}
	})
}

// validateIndexKind checks a new index request against the kind it names. Per-kind validation of
// the options and of the kinds of the fields being indexed belongs here.
func validateIndexKind(def client.CollectionVersion, req client.NewIndexRequest) error {
	if _, ok := indexKinds[req.Kind]; !ok {
		return NewErrUnknownIndexKind(req.Kind)
	}
	if req.Kind == client.IndexKindTrigram {
		return validateTrigramIndex(def, req)
	}
	return nil
}

// newLoadedCollectionIndex builds the write-path instance of an index read from a collection
// version.
//
// Collection versions are shared between peers, so a version can name an index kind this build
// does not implement. Such an index is recorded as failed and skipped rather than stopping the
// collection from opening; a nil instance with a nil error means it was skipped.
func (db *DB) newLoadedCollectionIndex(
	ctx context.Context,
	col client.Collection,
	desc client.IndexDescription,
	building bool,
) (client.CollectionIndex, error) {
	colIndex, err := NewCollectionIndex(ctx, col, desc, building)
	if errors.Is(err, ErrUnknownIndexKind) {
		if markErr := db.markIndexFailed(ctx, col.Version(), desc, err); markErr != nil {
			log.ErrorE("failed to record unknown index kind", errors.Join(markErr, err),
				corelog.String("collectionID", col.Version().CollectionID),
				corelog.Any("indexID", desc.ID),
			)
		}
		return nil, nil
	}
	return colIndex, err
}
