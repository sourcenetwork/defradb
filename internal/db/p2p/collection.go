// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"context"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/keys"
)

const marker = byte(0xff)

func (p *P2P) AddP2PCollections(
	ctx context.Context,
	collectionNames ...string,
) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	clientTxn := datastore.CtxMustGetClientTxn(ctx)
	txn := datastore.MustGetFromClientTxn(clientTxn)

	// first let's make sure the collections actually exists
	storeCollections := []client.Collection{}
	ident := identity.FromContext(ctx)
	for _, col := range collectionNames {
		storeCol, err := clientTxn.GetCollections(
			ctx,
			options.WithIdentity(options.GetCollections(), ident).SetCollectionName(col),
		)
		if err != nil {
			return err
		}
		if len(storeCol) == 0 {
			return client.NewErrCollectionNotFoundForName(col)
		}
		storeCollections = append(storeCollections, storeCol...)
	}

	// Ensure we can add all the collections to the store on the transaction
	// before adding to topics.
	for _, col := range storeCollections {
		key := keys.NewP2PCollectionKey(col.CollectionID())
		err := txn.Systemstore().Set(ctx, key.Bytes(), []byte{marker})
		if err != nil {
			return err
		}
	}

	txn.OnSuccessAsync(func() {
		for _, col := range storeCollections {
			err := p.host.AddPubSubTopic(col.CollectionID(), true, p.pubSubMessageHandler, p.peerEventHandler)
			if err != nil {
				log.ErrorE("Failed to add pubsub topic.", err)
			}
		}
	})

	return nil
}

func (p *P2P) DeleteP2PCollections(
	ctx context.Context,
	collectionNames ...string,
) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	clientTxn := datastore.CtxMustGetClientTxn(ctx)
	txn := datastore.MustGetFromClientTxn(clientTxn)

	// first let's make sure the collections actually exists
	storeCollections := []client.Collection{}
	ident := identity.FromContext(ctx)
	for _, col := range collectionNames {
		storeCol, err := clientTxn.GetCollections(
			ctx,
			options.WithIdentity(options.GetCollections(), ident).SetCollectionName(col),
		)
		if err != nil {
			return err
		}
		if len(storeCol) == 0 {
			return client.NewErrCollectionNotFoundForName(col)
		}
		storeCollections = append(storeCollections, storeCol...)
	}

	// Ensure we can remove all the collections to the store on the transaction
	// before adding to topics.
	for _, col := range storeCollections {
		key := keys.NewP2PCollectionKey(col.CollectionID())
		err := txn.Systemstore().Delete(ctx, key.Bytes())
		if err != nil {
			return err
		}
	}

	txn.OnSuccessAsync(func() {
		for _, col := range storeCollections {
			err := p.host.RemovePubSubTopic(col.CollectionID())
			if err != nil {
				log.ErrorE("Failed to remove pubsub topic.", err)
			}
		}
	})

	return nil
}

func (p *P2P) ListP2PCollections(
	ctx context.Context,
) ([]string, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	clientTxn := datastore.CtxMustGetClientTxn(ctx)
	txn := datastore.MustGetFromClientTxn(clientTxn)

	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix:   keys.NewP2PCollectionKey("").Bytes(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	collectionNames := []string{}
	ident := identity.FromContext(ctx)
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

		storeCol, err := clientTxn.GetCollections(
			ctx,
			options.WithIdentity(options.GetCollections(), ident).SetCollectionID(key.CollectionID),
		)
		if err != nil {
			return nil, err
		}
		if len(storeCol) == 0 {
			return nil, client.NewErrCollectionNotFoundForSchema(key.CollectionID)
		}
		collectionNames = append(collectionNames, storeCol[0].Name())
	}

	return collectionNames, iter.Close()
}

func (p *P2P) getAllP2PCollectionIDs(ctx context.Context) ([]string, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	iter, err := p.db.Multistore().Systemstore().Iterator(ctx, corekv.IterOptions{
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

func (p *P2P) loadAndPublishP2PCollections(ctx context.Context) error {
	collectionIDs, err := p.getAllP2PCollectionIDs(ctx)
	if err != nil {
		return err
	}
	for _, id := range collectionIDs {
		err := p.host.AddPubSubTopic(id, true, p.pubSubMessageHandler, p.peerEventHandler)
		if err != nil {
			return err
		}
	}

	return nil
}
