// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package lens

import (
	"context"

	"github.com/sourcenetwork/immutable/enumerable"
	"github.com/sourcenetwork/lens/host-go/config/model"
	"github.com/sourcenetwork/lens/host-go/engine/module"
	"github.com/sourcenetwork/lens/host-go/repository"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/description"
)

// DefaultPoolSize is the default size of the lens pool for each schema version.
const DefaultPoolSize int = 5

type LensRegistry struct {
	repository repository.Repository
}

var _ client.LensRegistry = (*LensRegistry)(nil)

func NewRegistry(
	poolSize int,
	runtime module.Runtime,
) *LensRegistry {
	return &LensRegistry{
		repository: repository.NewRepository(poolSize, runtime),
	}
}

func (r *LensRegistry) Init(txnSource client.TxnSource) {
	r.repository.Init(wrapSource(txnSource))
}

func (r *LensRegistry) SetMigration(ctx context.Context, collectionID string, cfg model.Lens) error {
	return r.repository.Add(ctx, collectionID, cfg)
}

func (r *LensRegistry) ReloadLenses(ctx context.Context) error {
	cols, err := description.GetCollections(ctx)
	if err != nil {
		return err
	}

	for _, col := range cols {
		sources := col.CollectionSources()

		if len(sources) == 0 {
			continue
		}

		// WARNING: Here we are only dealing with the first source in the set, this is fine for now as
		// currently collections can only have one source, however this code will need to change if/when
		// collections support multiple sources.

		if !sources[0].Transform.HasValue() {
			continue
		}

		err = r.SetMigration(ctx, col.VersionID, sources[0].Transform.Value())
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *LensRegistry) MigrateUp(
	ctx context.Context,
	src enumerable.Enumerable[LensDoc],
	collectionID string,
) (enumerable.Enumerable[map[string]any], error) {
	return r.repository.Transform(ctx, src, collectionID)
}

func (r *LensRegistry) MigrateDown(
	ctx context.Context,
	src enumerable.Enumerable[LensDoc],
	collectionID string,
) (enumerable.Enumerable[map[string]any], error) {
	return r.repository.Inverse(ctx, src, collectionID)
}

type txnSource struct {
	txnSource client.TxnSource
}

var _ repository.TxnSource = (*txnSource)(nil)

func wrapSource(s client.TxnSource) *txnSource {
	return &txnSource{
		txnSource: s,
	}
}

func (s *txnSource) NewTxn(ctx context.Context, readOnly bool) (repository.Txn, error) {
	txn, err := s.txnSource.NewTxn(ctx, readOnly)
	if err != nil {
		return nil, err
	}

	return datastore.MustGetFromClientTxn(txn), nil
}
