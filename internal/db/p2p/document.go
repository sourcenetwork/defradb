// Copyright 2025 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func (p *P2P) AddP2PDocuments(ctx context.Context, docIDs ...string) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	clientTxn := datastore.CtxMustGetClientTxn(ctx)
	txn := datastore.MustGetFromClientTxn(clientTxn)

	// Ensure we can add all the docIDs to the store on the transaction
	// before adding to topics.
	for _, docID := range docIDs {
		// ensure that the docID is a real docID.
		_, err := client.NewDocIDFromString(docID)
		if err != nil {
			return err
		}
		key := keys.NewP2PDocumentKey(docID)
		err = txn.Systemstore().Set(ctx, key.Bytes(), []byte{marker})
		if err != nil {
			return err
		}
	}

	txn.OnSuccess(func() {
		for _, docID := range docIDs {
			err := p.host.AddPubSubTopic(docID, true, p.pubSubMessageHandler)
			if err != nil {
				log.ErrorE("Failed to add pubsub topic.", err)
			}
		}
	})

	return nil
}

func (p *P2P) RemoveP2PDocuments(ctx context.Context, docIDs ...string) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	clientTxn := datastore.CtxMustGetClientTxn(ctx)
	txn := datastore.MustGetFromClientTxn(clientTxn)

	for _, docID := range docIDs {
		// ensure that the docID is a real docID.
		_, err := client.NewDocIDFromString(docID)
		if err != nil {
			return err
		}
		key := keys.NewP2PDocumentKey(docID)
		err = txn.Systemstore().Delete(ctx, key.Bytes())
		if err != nil {
			return err
		}
	}

	txn.OnSuccess(func() {
		for _, docID := range docIDs {
			err := p.host.RemovePubSubTopic(docID)
			if err != nil {
				log.ErrorE("Failed to remove pubsub topic.", err)
			}
		}
	})

	return nil
}

func (p *P2P) GetAllP2PDocuments(ctx context.Context) ([]string, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	clientTxn := datastore.CtxMustGetClientTxn(ctx)
	txn := datastore.MustGetFromClientTxn(clientTxn)

	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix:   keys.NewP2PDocumentKey("").Bytes(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	docIDs := []string{}
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}

		key, err := keys.NewP2PDocumentKeyFromString(string(iter.Key()))
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		docIDs = append(docIDs, key.DocID)
	}

	return docIDs, iter.Close()
}

func (p *P2P) loadAndPublishP2PDocuments(ctx context.Context) error {
	clientTxn, err := p.db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer clientTxn.Discard(ctx)
	ctx = datastore.CtxSetFromClientTxn(ctx, clientTxn)

	docIDs, err := p.GetAllP2PDocuments(ctx)
	if err != nil {
		return err
	}
	for _, docID := range docIDs {
		err := p.host.AddPubSubTopic(docID, true, p.pubSubMessageHandler)
		if err != nil {
			return err
		}
	}
	return nil
}
