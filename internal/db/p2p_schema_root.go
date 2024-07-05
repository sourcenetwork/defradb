// Copyright 2024 Democratized Data Foundation
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

	dsq "github.com/ipfs/go-datastore/query"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
)

const marker = byte(0xff)

func (db *db) AddP2PCollections(ctx context.Context, collectionIDs []string) error {
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	// TODO-ACP: Support ACP <> P2P - https://github.com/sourcenetwork/defradb/issues/2366
	// ctx = db.SetContextIdentity(ctx, identity)
	ctx = SetContextTxn(ctx, txn)

	// first let's make sure the collections actually exists
	storeCollections := []client.Collection{}
	for _, col := range collectionIDs {
		storeCol, err := db.GetCollections(
			ctx,
			client.CollectionFetchOptions{
				SchemaRoot: immutable.Some(col),
			},
		)
		if err != nil {
			return err
		}
		if len(storeCol) == 0 {
			return client.NewErrCollectionNotFoundForSchema(col)
		}
		storeCollections = append(storeCollections, storeCol...)
	}

	// Ensure none of the collections have a policy on them, until following is implemented:
	// TODO-ACP: ACP <> P2P https://github.com/sourcenetwork/defradb/issues/2366
	for _, col := range storeCollections {
		if col.Description().Policy.HasValue() {
			return ErrP2PColHasPolicy
		}
	}

	evt := event.P2PTopic{}

	// Ensure we can add all the collections to the store on the transaction
	// before adding to topics.
	for _, col := range storeCollections {
		key := core.NewP2PCollectionKey(col.SchemaRoot())
		err = txn.Systemstore().Put(ctx, key.ToDS(), []byte{marker})
		if err != nil {
			return err
		}
		evt.ToAdd = append(evt.ToAdd, col.SchemaRoot())
	}

	for _, col := range storeCollections {
		keyChan, err := col.GetAllDocIDs(ctx)
		if err != nil {
			return err
		}
		for key := range keyChan {
			evt.ToRemove = append(evt.ToRemove, key.ID.String())
		}
	}

	txn.OnSuccess(func() {
		db.events.Publish(event.NewMessage(event.P2PTopicName, evt))
	})

	return txn.Commit(ctx)
}

func (db *db) RemoveP2PCollections(ctx context.Context, collectionIDs []string) error {
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	// TODO-ACP: Support ACP <> P2P - https://github.com/sourcenetwork/defradb/issues/2366
	// ctx = db.SetContextIdentity(ctx, identity)
	ctx = SetContextTxn(ctx, txn)

	// first let's make sure the collections actually exists
	storeCollections := []client.Collection{}
	for _, col := range collectionIDs {
		storeCol, err := db.GetCollections(
			ctx,
			client.CollectionFetchOptions{
				SchemaRoot: immutable.Some(col),
			},
		)
		if err != nil {
			return err
		}
		if len(storeCol) == 0 {
			return client.NewErrCollectionNotFoundForSchema(col)
		}
		storeCollections = append(storeCollections, storeCol...)
	}

	evt := event.P2PTopic{}

	// Ensure we can remove all the collections to the store on the transaction
	// before adding to topics.
	for _, col := range storeCollections {
		key := core.NewP2PCollectionKey(col.SchemaRoot())
		err = txn.Systemstore().Delete(ctx, key.ToDS())
		if err != nil {
			return err
		}
		evt.ToRemove = append(evt.ToRemove, col.SchemaRoot())
	}

	for _, col := range storeCollections {
		keyChan, err := col.GetAllDocIDs(ctx)
		if err != nil {
			return err
		}
		for key := range keyChan {
			evt.ToAdd = append(evt.ToAdd, key.ID.String())
		}
	}

	txn.OnSuccess(func() {
		db.events.Publish(event.NewMessage(event.P2PTopicName, evt))
	})

	return txn.Commit(ctx)
}

func (db *db) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	txn, err := db.NewTxn(ctx, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	query := dsq.Query{
		Prefix: core.NewP2PCollectionKey("").ToString(),
	}
	results, err := txn.Systemstore().Query(ctx, query)
	if err != nil {
		return nil, err
	}

	collectionIDs := []string{}
	for result := range results.Next() {
		key, err := core.NewP2PCollectionKeyFromString(result.Key)
		if err != nil {
			return nil, err
		}
		collectionIDs = append(collectionIDs, key.CollectionID)
	}

	return collectionIDs, nil
}

func (db *db) PeerInfo() peer.AddrInfo {
	peerInfo := db.peerInfo.Load()
	if peerInfo != nil {
		return peerInfo.(peer.AddrInfo)
	}
	return peer.AddrInfo{}
}

func (db *db) loadAndPublishP2PCollections(ctx context.Context) error {
	schemaRoots, err := db.GetAllP2PCollections(ctx)
	if err != nil {
		return err
	}
	db.events.Publish(event.NewMessage(event.P2PTopicName, event.P2PTopic{
		ToAdd: schemaRoots,
	}))

	// Get all DocIDs across all collections in the DB
	cols, err := db.GetCollections(ctx, client.CollectionFetchOptions{})
	if err != nil {
		return err
	}

	// Index the schema roots for faster lookup.
	colMap := make(map[string]struct{})
	for _, schemaRoot := range schemaRoots {
		colMap[schemaRoot] = struct{}{}
	}

	for _, col := range cols {
		// If we subscribed to the collection, we skip subscribing to the collection's docIDs.
		if _, ok := colMap[col.SchemaRoot()]; ok {
			continue
		}
		// TODO-ACP: Support ACP <> P2P - https://github.com/sourcenetwork/defradb/issues/2366
		docIDChan, err := col.GetAllDocIDs(ctx)
		if err != nil {
			return err
		}

		for docID := range docIDChan {
			db.events.Publish(event.NewMessage(event.P2PTopicName, event.P2PTopic{
				ToAdd: []string{docID.ID.String()},
			}))
		}
	}
	return nil
}
