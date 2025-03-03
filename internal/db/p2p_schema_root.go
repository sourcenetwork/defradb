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

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/keys"
)

const marker = byte(0xff)

func (db *DB) AddP2PCollections(ctx context.Context, collectionIDs []string) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

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

	if db.acp.HasValue() && !db.acp.Value().SupportsP2P() {
		for _, col := range storeCollections {
			if col.Description().Policy.HasValue() {
				return ErrP2PColHasPolicy
			}
		}
	}

	evt := event.P2PTopic{}

	// Ensure we can add all the collections to the store on the transaction
	// before adding to topics.
	for _, col := range storeCollections {
		key := keys.NewP2PCollectionKey(col.SchemaRoot())
		err = txn.Systemstore().Set(ctx, key.Bytes(), []byte{marker})
		if err != nil {
			return err
		}
		evt.ToAdd = append(evt.ToAdd, col.SchemaRoot())
	}

	// This is a node specific action which means the actor is the node itself.
	ctx = identity.WithContext(ctx, db.nodeIdentity)
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

func (db *DB) RemoveP2PCollections(ctx context.Context, collectionIDs []string) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

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
		key := keys.NewP2PCollectionKey(col.SchemaRoot())
		err = txn.Systemstore().Delete(ctx, key.Bytes())
		if err != nil {
			return err
		}
		evt.ToRemove = append(evt.ToRemove, col.SchemaRoot())
	}

	// This is a node specific action which means the actor is the node itself.
	ctx = identity.WithContext(ctx, db.nodeIdentity)
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

func (db *DB) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	txn, err := db.NewTxn(ctx, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix:   keys.NewP2PCollectionKey("").Bytes(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	collectionIDs := []string{}
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}

		key, err := keys.NewP2PCollectionKeyFromString(string(iter.Key()))
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		collectionIDs = append(collectionIDs, key.CollectionID)
	}

	return collectionIDs, iter.Close()
}

func (db *DB) PeerInfo() peer.AddrInfo {
	peerInfo := db.peerInfo.Load()
	if peerInfo != nil {
		return peerInfo.(peer.AddrInfo)
	}
	return peer.AddrInfo{}
}

func (db *DB) loadAndPublishP2PCollections(ctx context.Context) error {
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

	// This is a node specific action which means the actor is the node itself.
	ctx = identity.WithContext(ctx, db.nodeIdentity)
	for _, col := range cols {
		// If we subscribed to the collection, we skip subscribing to the collection's docIDs.
		if _, ok := colMap[col.SchemaRoot()]; ok {
			continue
		}
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
