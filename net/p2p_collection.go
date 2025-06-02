// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import (
	"context"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/keys"
)

const marker = byte(0xff)

func (p *Peer) AddP2PCollections(ctx context.Context, collectionIDs ...string) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn := datastore.EnsureContextTxn(ctx, p.db.Rootstore(), false)
	defer txn.Discard(ctx)

	// first let's make sure the collections actually exists
	storeCollections := []client.Collection{}
	for _, col := range collectionIDs {
		storeCol, err := p.db.GetCollections(
			ctx,
			client.CollectionFetchOptions{
				CollectionID: immutable.Some(col),
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

	// Ensure we can add all the collections to the store on the transaction
	// before adding to topics.
	for _, col := range storeCollections {
		key := keys.NewP2PCollectionKey(col.SchemaRoot())
		err := datastore.SystemstoreFrom(txn.Store()).Set(ctx, key.Bytes(), []byte{marker})
		if err != nil {
			return err
		}
	}

	txn.OnSuccess(func() {
		for _, col := range storeCollections {
			_, err := p.server.addPubSubTopic(col.SchemaRoot(), true, nil)
			if err != nil {
				log.ErrorE("Failed to add pubsub topic.", err)
			}
		}
	})

	return txn.Commit(ctx)
}

func (p *Peer) RemoveP2PCollections(ctx context.Context, collectionIDs ...string) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn := datastore.EnsureContextTxn(ctx, p.db.Rootstore(), false)
	defer txn.Discard(ctx)

	// first let's make sure the collections actually exists
	storeCollections := []client.Collection{}
	for _, col := range collectionIDs {
		storeCol, err := p.db.GetCollections(
			ctx,
			client.CollectionFetchOptions{
				CollectionID: immutable.Some(col),
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

	// Ensure we can remove all the collections to the store on the transaction
	// before adding to topics.
	for _, col := range storeCollections {
		key := keys.NewP2PCollectionKey(col.SchemaRoot())
		err := datastore.SystemstoreFrom(txn.Store()).Delete(ctx, key.Bytes())
		if err != nil {
			return err
		}
	}

	txn.OnSuccess(func() {
		for _, col := range storeCollections {
			err := p.server.removePubSubTopic(col.SchemaRoot())
			if err != nil {
				log.ErrorE("Failed to remove pubsub topic.", err)
			}
		}
	})

	return txn.Commit(ctx)
}

func (p *Peer) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn := datastore.EnsureContextTxn(ctx, p.db.Rootstore(), false)
	defer txn.Discard(ctx)

	iter, err := datastore.SystemstoreFrom(txn.Store()).Iterator(ctx, corekv.IterOptions{
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

func (p *Peer) loadAndPublishP2PCollections(ctx context.Context) error {
	collectionIDs, err := p.GetAllP2PCollections(ctx)
	if err != nil {
		return err
	}
	for _, id := range collectionIDs {
		_, err := p.server.addPubSubTopic(id, true, nil)
		if err != nil {
			return err
		}
	}

	// Get all DocIDs across all collections in the DB
	cols, err := p.db.GetCollections(ctx, client.CollectionFetchOptions{})
	if err != nil {
		return err
	}

	// Index the schema roots for faster lookup.
	colMap := make(map[string]struct{})
	for _, id := range collectionIDs {
		colMap[id] = struct{}{}
	}

	// This is a node specific action which means the actor is the node itself.
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
			_, err := p.server.addPubSubTopic(docID.ID.String(), true, nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
