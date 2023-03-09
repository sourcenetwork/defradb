// Copyright 2022 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
)

const marker = byte(0xff)

// AddP2PCollection adds the given collection ID that the P2P system
// subscribes to to the the persisted list. It will error if the provided
// collection ID is invalid.
func (db *innerDB) AddP2PCollection(ctx context.Context, collectionID string) error {
	txn, err := db.getTxn(ctx, false)
	if err != nil {
		return err
	}
	defer db.discardImplicitTxn(ctx, txn)

	err = db.addP2PCollectionTxn(ctx, txn, collectionID)
	if err != nil {
		return err
	}

	return db.commitImplicitTxn(ctx, txn)
}

// AddP2PCollectionTxn adds the given collection ID that the P2P system
// subscribes to to the the persisted list. It will error if the provided
// collection ID is invalid.
func (db *innerDB) addP2PCollectionTxn(ctx context.Context, txn datastore.Txn, collectionID string) error {
	_, err := db.getCollectionBySchemaIDTxn(ctx, txn, collectionID)
	if err != nil {
		return NewErrAddingP2PCollection(err)
	}
	key := core.NewP2PCollectionKey(collectionID)
	return txn.Systemstore().Put(ctx, key.ToDS(), []byte{marker})
}

// RemoveP2PCollection removes the given collection ID that the P2P system
// subscribes to from the the persisted list. It will error if the provided
// collection ID is invalid.
func (db *innerDB) RemoveP2PCollection(ctx context.Context, collectionID string) error {
	txn, err := db.getTxn(ctx, false)
	if err != nil {
		return err
	}
	defer db.discardImplicitTxn(ctx, txn)

	err = db.removeP2PCollectionTxn(ctx, txn, collectionID)
	if err != nil {
		return err
	}

	return db.commitImplicitTxn(ctx, txn)
}

// RemoveP2PCollectionTxn removes the given collection ID that the P2P system
// subscribes to from the the persisted list. It will error if the provided
// collection ID is invalid.
func (db *innerDB) removeP2PCollectionTxn(ctx context.Context, txn datastore.Txn, collectionID string) error {
	_, err := db.getCollectionBySchemaIDTxn(ctx, txn, collectionID)
	if err != nil {
		return NewErrRemovingP2PCollection(err)
	}
	key := core.NewP2PCollectionKey(collectionID)
	return txn.Systemstore().Delete(ctx, key.ToDS())
}

func (db *innerDB) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	txn, err := db.getTxn(ctx, false)
	if err != nil {
		return []string{}, err
	}
	defer db.discardImplicitTxn(ctx, txn)

	colIDs, err := db.getAllP2PCollectionsTxn(ctx, txn)
	if err != nil {
		return []string{}, err
	}

	err = db.commitImplicitTxn(ctx, txn)
	return colIDs, err
}

// GetAllP2PCollections returns the list of persisted collection IDs that
// the P2P system subscribes to.
func (db *innerDB) getAllP2PCollectionsTxn(ctx context.Context, txn datastore.Txn) ([]string, error) {
	prefix := core.NewP2PCollectionKey("")
	results, err := txn.Systemstore().Query(ctx, dsq.Query{
		Prefix: prefix.ToString(),
	})
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
