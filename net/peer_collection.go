package net

import (
	"context"

	dsq "github.com/ipfs/go-datastore/query"

	"github.com/sourcenetwork/defradb/core"
)

const marker = byte(0xff)

func (p *Peer) AddP2PCollection(ctx context.Context, collectionID string) error {
	txn, err := p.db.NewTxn(p.ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(p.ctx)

	// first let's make sure the collections actually exists
	collection, err := p.db.WithTxn(txn).GetCollectionBySchemaID(ctx, collectionID)
	if err != nil {
		return err
	}

	// Ensure we can add all the collections to the store on the transaction
	// before adding to topics.
	key := core.NewP2PCollectionKey(collectionID)
	err = txn.Systemstore().Put(ctx, key.ToDS(), []byte{marker})
	if err != nil {
		return err
	}

	// Add pubsub topics and remove them if we get an error.
	err = p.server.addPubSubTopic(collectionID, true)
	if err != nil {
		return p.rollbackAddPubSubTopics(err, collectionID)
	}

	keyChan, err := collection.WithTxn(txn).GetAllDocKeys(p.ctx)
	if err != nil {
		return err
	}

	// After adding the collection topics, we remove the collections' documents
	// from the pubsub topics to avoid receiving duplicate events.
	removedTopics := []string{}
	for res := range keyChan {
		err := p.server.removePubSubTopic(res.Key.String())
		if err != nil {
			return p.rollbackRemovePubSubTopics(err, removedTopics...)
		}
		removedTopics = append(removedTopics, res.Key.String())
	}

	if err = txn.Commit(p.ctx); err != nil {
		err = p.rollbackRemovePubSubTopics(err, removedTopics...)
		return p.rollbackAddPubSubTopics(err, collectionID)
	}

	return nil
}

func (p *Peer) RemoveP2PCollection(ctx context.Context, collectionID string) error {
	txn, err := p.db.NewTxn(p.ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(p.ctx)

	// first let's make sure the collections actually exists
	collection, err := p.db.WithTxn(txn).GetCollectionBySchemaID(ctx, collectionID)
	if err != nil {
		return err
	}

	// Ensure we can remove all the collections to the store on the transaction
	// before adding to topics.
	key := core.NewP2PCollectionKey(collectionID)
	err = txn.Systemstore().Delete(ctx, key.ToDS())
	if err != nil {
		return err
	}

	// Remove pubsub topics and add them back if we get an error.
	err = p.server.removePubSubTopic(collectionID)
	if err != nil {
		return p.rollbackRemovePubSubTopics(err, collectionID)
	}

	keyChan, err := collection.WithTxn(txn).GetAllDocKeys(p.ctx)
	if err != nil {
		return err
	}

	// After removing the collection topics, we add back the collections' documents
	// to the pubsub topics.
	addedTopics := []string{}
	for key := range keyChan {
		err := p.server.addPubSubTopic(key.Key.String(), true)
		if err != nil {
			return p.rollbackAddPubSubTopics(err, addedTopics...)
		}
		addedTopics = append(addedTopics, key.Key.String())
	}

	if err = txn.Commit(p.ctx); err != nil {
		err = p.rollbackAddPubSubTopics(err, addedTopics...)
		return p.rollbackRemovePubSubTopics(err, collectionID)
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
