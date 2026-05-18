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
	"testing"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/blockstore"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/dac"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// If a child fetcher fails to start mid-way, MultiVersioned still has to
// know about it so Close can clean it up. Otherwise the child is stranded
// with its in-memory store and any badger iterators it opened on the
// parent txn, and the next Discard on that txn panics with
// "Unclosed iterator at time of Txn.Discard". That's the actual crash
// the subscription handler hits in production when a merge fails part-way
// through a fetch.
//
// To force the failure we feed in a hand-crafted block whose delta names
// a field the schema doesn't have, so merge raises ErrFieldNotExist and
// Start bails out. The check is: did the child get tracked anyway?
func TestMultiVersioned_Start_OnChildStartError_TracksChildForClose(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	// Stash a block whose link references a field that doesn't exist in
	// the schema. merge will hit it and bail out.
	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(
		blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())),
	)

	doc, err := client.NewDocFromMap(
		ctx,
		map[string]any{"name": "Alice"},
		col.Version(),
	)
	require.NoError(t, err)
	docID := []byte(doc.ID().String())

	compositeLink := authorBlockWithUnknownField(
		t,
		&lsys,
		docID,
		col.Version().VersionID,
		"nonexistent_field",
	)

	// Use a real top-level badger txn — same as what the subscription
	// handler opens per event.
	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	defer txn.Discard()

	ctx = InitContext(ctx, txn)

	f := &fetcher.MultiVersioned{}
	err = f.Init(
		ctx,
		immutable.None[acpIdentity.Identity](),
		txn.(*Txn).BasicTxn,
		db.nodeACP,
		immutable.None[dac.DocumentACP](),
		immutable.None[client.IndexDescription](),
		col,
		col.Version().Fields,
		nil,
		nil,
		nil,
		false,
	)
	require.NoError(t, err)

	prefix := keys.HeadstoreDocKey{Cid: compositeLink.Cid}

	startErr := f.Start(ctx, prefix)
	require.Error(t, startErr, "Start should fail when the delta names a field the schema doesn't have")

	// This is the whole point of the test. Init succeeded, so the child
	// has resources to release; Close needs to be able to find it.
	require.Equal(
		t,
		1,
		f.NumTrackedChildrenForTest(),
		"child must be tracked after a Start failure so Close can release it",
	)

	require.NotPanics(t, func() {
		_ = f.Close()
	})
}

// Writes a field-delta block naming the given (unknown) field plus a
// composite block linking to it, and returns the composite's link.
func authorBlockWithUnknownField(
	t *testing.T,
	lsys *linking.LinkSystem,
	docID []byte,
	collectionVersionID string,
	unknownField string,
) cidlink.Link {
	t.Helper()

	fieldBlock := coreblock.Block{
		Delta: crdt.CRDT{
			LWWDelta: &crdt.LWWDelta{
				DocID:               docID,
				FieldName:           unknownField,
				Priority:            1,
				CollectionVersionID: collectionVersionID,
				Data:                encodeValue("bogus"),
			},
		},
	}
	fieldLink, err := lsys.Store(
		ipld.LinkContext{},
		coreblock.GetLinkPrototype(),
		fieldBlock.GenerateNode(),
	)
	require.NoError(t, err)

	compositeBlock := coreblock.New(
		crdt.NewCRDT(&crdt.DocCompositeDelta{
			DocID:               docID,
			Priority:            1,
			CollectionVersionID: collectionVersionID,
			Status:              1,
		}),
		[]coreblock.DAGLink{
			{Name: unknownField, Link: fieldLink.(cidlink.Link)},
		},
	)
	compositeLink, err := lsys.Store(
		ipld.LinkContext{},
		coreblock.GetLinkPrototype(),
		compositeBlock.GenerateNode(),
	)
	require.NoError(t, err)

	return compositeLink.(cidlink.Link)
}
