// Copyright 2023 Democratized Data Foundation
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

	dsq "github.com/ipfs/go-datastore/query"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
)

const marker = byte(0xff)

func (p *Peer) AddP2PCollections(ctx context.Context, collectionIDs []string) error {
	txn, err := p.db.NewTxn(p.ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(p.ctx)

	// first let's make sure the collections actually exists
	storeCollections := []client.Collection{}
	for _, col := range collectionIDs {
		storeCol, err := p.db.WithTxn(txn).GetCollectionBySchemaID(p.ctx, col)
		if err != nil {
			return err
		}
		storeCollections = append(storeCollections, storeCol)
	}

	// Ensure we can add all the collections to the store on the transaction
	// before adding to topics.
	for _, col := range storeCollections {
		key := core.NewP2PCollectionKey(col.SchemaID())
		err = txn.Systemstore().Put(ctx, key.ToDS(), []byte{marker})
		if err != nil {
			return err
		}
	}

	// Add pubsub topics and remove them if we get an error.
	addedTopics := []string{}
	for _, col := range collectionIDs {
		err = p.server.addPubSubTopic(col, true)
		if err != nil {
			return p.rollbackAddPubSubTopics(addedTopics, err)
		}
		addedTopics = append(addedTopics, col)
	}

	// After adding the collection topics, we remove the collections' documents
	// from the pubsub topics to avoid receiving duplicate events.
	removedTopics := []string{}
	for _, col := range storeCollections {
		keyChan, err := col.GetAllDocKeys(p.ctx)
		if err != nil {
			return err
		}
		for key := range keyChan {
			err := p.server.removePubSubTopic(key.Key.String())
			if err != nil {
				return p.rollbackRemovePubSubTopics(removedTopics, err)
			}
			removedTopics = append(removedTopics, key.Key.String())
		}
	}

	if err = txn.Commit(p.ctx); err != nil {
		err = p.rollbackRemovePubSubTopics(removedTopics, err)
		return p.rollbackAddPubSubTopics(addedTopics, err)
	}

	return nil
}

func (p *Peer) RemoveP2PCollections(ctx context.Context, collectionIDs []string) error {
	txn, err := p.db.NewTxn(p.ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(p.ctx)

	// first let's make sure the collections actually exists
	storeCollections := []client.Collection{}
	for _, col := range collectionIDs {
		storeCol, err := p.db.WithTxn(txn).GetCollectionBySchemaID(p.ctx, col)
		if err != nil {
			return err
		}
		storeCollections = append(storeCollections, storeCol)
	}

	// Ensure we can remove all the collections to the store on the transaction
	// before adding to topics.
	for _, col := range storeCollections {
		key := core.NewP2PCollectionKey(col.SchemaID())
		err = txn.Systemstore().Delete(ctx, key.ToDS())
		if err != nil {
			return err
		}
	}

	// Remove pubsub topics and add them back if we get an error.
	removedTopics := []string{}
	for _, col := range collectionIDs {
		err = p.server.removePubSubTopic(col)
		if err != nil {
			return p.rollbackRemovePubSubTopics(removedTopics, err)
		}
		removedTopics = append(removedTopics, col)
	}

	// After removing the collection topics, we add back the collections' documents
	// to the pubsub topics.
	addedTopics := []string{}
	for _, col := range storeCollections {
		keyChan, err := col.GetAllDocKeys(p.ctx)
		if err != nil {
			return err
		}
		for key := range keyChan {
			err := p.server.addPubSubTopic(key.Key.String(), true)
			if err != nil {
				return p.rollbackAddPubSubTopics(addedTopics, err)
			}
			addedTopics = append(addedTopics, key.Key.String())
		}
	}

	if err = txn.Commit(p.ctx); err != nil {
		err = p.rollbackAddPubSubTopics(addedTopics, err)
		return p.rollbackRemovePubSubTopics(removedTopics, err)
	}

	return nil
}

func (p *Peer) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	txn, err := p.db.NewTxn(p.ctx, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(p.ctx)

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
