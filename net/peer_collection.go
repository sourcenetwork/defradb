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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/internal/core"
)

const marker = byte(0xff)

func (p *Peer) AddP2PCollections(ctx context.Context, collectionIDs []string) error {
	txn, err := p.db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	// TODO-ACP: Support ACP <> P2P - https://github.com/sourcenetwork/defradb/issues/2366
	// ctx = db.SetContextIdentity(ctx, identity)
	ctx = db.SetContextTxn(ctx, txn)

	// first let's make sure the collections actually exists
	storeCollections := []client.Collection{}
	for _, col := range collectionIDs {
		storeCol, err := p.db.GetCollections(
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

	// Ensure we can add all the collections to the store on the transaction
	// before adding to topics.
	for _, col := range storeCollections {
		key := core.NewP2PCollectionKey(col.SchemaRoot())
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
		keyChan, err := col.GetAllDocIDs(ctx)
		if err != nil {
			return err
		}
		for key := range keyChan {
			err := p.server.removePubSubTopic(key.ID.String())
			if err != nil {
				return p.rollbackRemovePubSubTopics(removedTopics, err)
			}
			removedTopics = append(removedTopics, key.ID.String())
		}
	}

	if err = txn.Commit(ctx); err != nil {
		err = p.rollbackRemovePubSubTopics(removedTopics, err)
		return p.rollbackAddPubSubTopics(addedTopics, err)
	}

	return nil
}

func (p *Peer) RemoveP2PCollections(ctx context.Context, collectionIDs []string) error {
	txn, err := p.db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	// TODO-ACP: Support ACP <> P2P - https://github.com/sourcenetwork/defradb/issues/2366
	// ctx = db.SetContextIdentity(ctx, identity)
	ctx = db.SetContextTxn(ctx, txn)

	// first let's make sure the collections actually exists
	storeCollections := []client.Collection{}
	for _, col := range collectionIDs {
		storeCol, err := p.db.GetCollections(
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

	// Ensure we can remove all the collections to the store on the transaction
	// before adding to topics.
	for _, col := range storeCollections {
		key := core.NewP2PCollectionKey(col.SchemaRoot())
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
		keyChan, err := col.GetAllDocIDs(ctx)
		if err != nil {
			return err
		}
		for key := range keyChan {
			err := p.server.addPubSubTopic(key.ID.String(), true)
			if err != nil {
				return p.rollbackAddPubSubTopics(addedTopics, err)
			}
			addedTopics = append(addedTopics, key.ID.String())
		}
	}

	if err = txn.Commit(ctx); err != nil {
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
