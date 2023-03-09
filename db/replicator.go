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
	"encoding/json"
	"errors"

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
)

// SetReplicator adds a new replicator to the database.
func (db *innerDB) SetReplicator(ctx context.Context, rep client.Replicator) error {
	txn, err := db.getTxn(ctx, false)
	if err != nil {
		return err
	}
	defer db.discardImplicitTxn(ctx, txn)

	err = db.setReplicatorTxn(ctx, txn, rep)
	if err != nil {
		return err
	}

	return db.commitImplicitTxn(ctx, txn)
}

func (db *innerDB) setReplicatorTxn(ctx context.Context, txn datastore.Txn, rep client.Replicator) error {
	existingRep, err := db.getReplicatorTxn(ctx, txn, rep.Info)
	if errors.Is(err, ds.ErrNotFound) {
		return db.saveReplicatorTxn(ctx, txn, rep)
	}
	if err != nil {
		return err
	}

	newSchemas := []string{}
	for _, newSchema := range rep.Schemas {
		isNew := true
		for _, existingSchema := range existingRep.Schemas {
			if existingSchema == newSchema {
				isNew = false
				break
			}
		}
		if isNew {
			newSchemas = append(newSchemas, newSchema)
		}
	}
	rep.Schemas = append(existingRep.Schemas, newSchemas...)
	return db.saveReplicatorTxn(ctx, txn, rep)
}

// DeleteReplicator removes a replicator from the database.
func (db *innerDB) DeleteReplicator(ctx context.Context, rep client.Replicator) error {
	txn, err := db.getTxn(ctx, false)
	if err != nil {
		return err
	}
	defer db.discardImplicitTxn(ctx, txn)

	err = db.deleteReplicatorWithSchemas(ctx, txn, rep)
	if err != nil {
		return err
	}

	err = db.commitImplicitTxn(ctx, txn)
	return err
}

func (db *innerDB) deleteReplicatorWithSchemas(ctx context.Context, txn datastore.Txn, rep client.Replicator) error {
	if len(rep.Schemas) == 0 {
		return db.deleteReplicatorTxn(ctx, txn, rep.Info.ID)
	}
	return db.deleteSchemasForReplicatorTxn(ctx, txn, rep)
}

func (db *innerDB) deleteReplicatorTxn(ctx context.Context, txn datastore.Txn, pid peer.ID) error {
	key := core.NewReplicatorKey(pid.String())
	return txn.Systemstore().Delete(ctx, key.ToDS())
}

func (db *innerDB) deleteSchemasForReplicatorTxn(ctx context.Context, txn datastore.Txn, rep client.Replicator) error {
	existingRep, err := db.getReplicatorTxn(ctx, txn, rep.Info)
	if err != nil {
		return err
	}

	updatedSchemaList := []string{}
	for _, s := range existingRep.Schemas {
		found := false
		for _, toDelete := range rep.Schemas {
			if toDelete == s {
				found = true
				break
			}
		}
		if !found {
			updatedSchemaList = append(updatedSchemaList, s)
		}
	}

	if len(updatedSchemaList) == 0 {
		return db.deleteReplicatorTxn(ctx, txn, rep.Info.ID)
	}

	existingRep.Schemas = updatedSchemaList
	return db.saveReplicatorTxn(ctx, txn, existingRep)
}

// GetAllReplicators returns all replicators of the database.
func (db *innerDB) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	txn, err := db.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer db.discardImplicitTxn(ctx, txn)

	reps, err := db.getAllReplicatorsTxn(ctx, txn)
	if err != nil {
		return nil, err
	}

	err = db.commitImplicitTxn(ctx, txn)
	return reps, err
}

func (db *innerDB) getAllReplicatorsTxn(ctx context.Context, txn datastore.Txn) ([]client.Replicator, error) {
	reps := []client.Replicator{}
	// create collection system prefix query
	prefix := core.NewReplicatorKey("")
	results, err := txn.Systemstore().Query(ctx, dsq.Query{
		Prefix: prefix.ToString(),
	})
	if err != nil {
		return nil, err
	}

	for result := range results.Next() {
		var rep client.Replicator
		err = json.Unmarshal(result.Value, &rep)
		if err != nil {
			return nil, err
		}

		reps = append(reps, rep)
	}

	return reps, nil
}

// GetReplicator
func (db *innerDB) GetReplicator(ctx context.Context, info peer.AddrInfo) (client.Replicator, error) {
	txn, err := db.getTxn(ctx, false)
	if err != nil {
		return client.Replicator{}, err
	}
	defer db.discardImplicitTxn(ctx, txn)

	rep, err := db.getReplicatorTxn(ctx, txn, info)
	if err != nil {
		return client.Replicator{}, err
	}

	err = db.commitImplicitTxn(ctx, txn)
	return rep, err
}

func (db *innerDB) getReplicatorTxn(
	ctx context.Context,
	txn datastore.Txn,
	info peer.AddrInfo,
) (client.Replicator, error) {
	rep := client.Replicator{}
	key := core.NewReplicatorKey(info.ID.String())
	value, err := txn.Systemstore().Get(ctx, key.ToDS())
	if err != nil {
		return rep, err
	}

	err = json.Unmarshal(value, &rep)
	if err != nil {
		return rep, err
	}

	return rep, nil
}

func (db *innerDB) saveReplicatorTxn(ctx context.Context, txn datastore.Txn, rep client.Replicator) error {
	key := core.NewReplicatorKey(rep.Info.ID.String())
	repBytes, err := json.Marshal(rep)
	if err != nil {
		return err
	}
	return txn.Systemstore().Put(ctx, key.ToDS(), repBytes)
}
