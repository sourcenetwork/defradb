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
	"encoding/json"

	dsq "github.com/ipfs/go-datastore/query"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db"
)

func (p *Peer) SetReplicator(ctx context.Context, rep client.Replicator) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	txn, err := p.db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	if rep.Info.ID == p.host.ID() {
		return ErrSelfTargetForReplicator
	}
	if err := rep.Info.ID.Validate(); err != nil {
		return err
	}

	// use a session for all operations
	sess := db.NewSession(ctx).WithTxn(txn)

	var collections []client.Collection
	switch {
	case len(rep.Schemas) > 0:
		// if specific collections are chosen get them by name
		for _, name := range rep.Schemas {
			col, err := p.db.GetCollectionByName(sess, name)
			if err != nil {
				return NewErrReplicatorCollections(err)
			}

			if col.Description().Policy.HasValue() {
				return ErrReplicatorColHasPolicy
			}

			collections = append(collections, col)
		}

	default:
		// default to all collections (unless a collection contains a policy).
		// TODO-ACP: default to all collections after resolving https://github.com/sourcenetwork/defradb/issues/2366
		allCollections, err := p.db.GetCollections(sess, client.CollectionFetchOptions{})
		if err != nil {
			return NewErrReplicatorCollections(err)
		}

		for _, col := range allCollections {
			// Can not default to all collections if any collection has a policy.
			// TODO-ACP: remove this check/loop after https://github.com/sourcenetwork/defradb/issues/2366
			if col.Description().Policy.HasValue() {
				return ErrReplicatorSomeColsHavePolicy
			}
		}
		collections = allCollections
	}
	rep.Schemas = nil

	// Add the destination's peer multiaddress in the peerstore.
	// This will be used during connection and stream creation by libp2p.
	p.host.Peerstore().AddAddrs(rep.Info.ID, rep.Info.Addrs, peerstore.PermanentAddrTTL)

	var added []client.Collection
	for _, col := range collections {
		reps, exists := p.replicators[col.SchemaRoot()]
		if !exists {
			p.replicators[col.SchemaRoot()] = make(map[peer.ID]struct{})
		}
		if _, exists := reps[rep.Info.ID]; !exists {
			// keep track of newly added collections so we don't
			// push logs to a replicator peer multiple times.
			p.replicators[col.SchemaRoot()][rep.Info.ID] = struct{}{}
			added = append(added, col)
		}
		rep.Schemas = append(rep.Schemas, col.SchemaRoot())
	}

	// persist replicator to the datastore
	repBytes, err := json.Marshal(rep)
	if err != nil {
		return err
	}
	key := core.NewReplicatorKey(rep.Info.ID.String())
	err = txn.Systemstore().Put(ctx, key.ToDS(), repBytes)
	if err != nil {
		return err
	}

	// push all collection documents to the replicator peer
	for _, col := range added {
		// TODO-ACP: Support ACP <> P2P - https://github.com/sourcenetwork/defradb/issues/2366
		keysCh, err := col.GetAllDocIDs(sess, acpIdentity.NoIdentity)
		if err != nil {
			return NewErrReplicatorDocID(err, col.Name().Value(), rep.Info.ID)
		}
		p.pushToReplicator(ctx, txn, col, keysCh, rep.Info.ID)
	}

	return txn.Commit(ctx)
}

func (p *Peer) DeleteReplicator(ctx context.Context, rep client.Replicator) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	txn, err := p.db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	if rep.Info.ID == p.host.ID() {
		return ErrSelfTargetForReplicator
	}
	if err := rep.Info.ID.Validate(); err != nil {
		return err
	}

	// use a session for all operations
	sess := db.NewSession(ctx).WithTxn(txn)

	var collections []client.Collection
	switch {
	case len(rep.Schemas) > 0:
		// if specific collections are chosen get them by name
		for _, name := range rep.Schemas {
			col, err := p.db.GetCollectionByName(sess, name)
			if err != nil {
				return NewErrReplicatorCollections(err)
			}
			collections = append(collections, col)
		}
		// make sure the replicator exists in the datastore
		key := core.NewReplicatorKey(rep.Info.ID.String())
		_, err = txn.Systemstore().Get(ctx, key.ToDS())
		if err != nil {
			return err
		}

	default:
		// default to all collections
		collections, err = p.db.GetCollections(sess, client.CollectionFetchOptions{})
		if err != nil {
			return NewErrReplicatorCollections(err)
		}
	}
	rep.Schemas = nil

	schemaMap := make(map[string]struct{})
	for _, col := range collections {
		schemaMap[col.SchemaRoot()] = struct{}{}
	}

	// update replicators and add remaining schemas to rep
	for key, val := range p.replicators {
		if _, exists := val[rep.Info.ID]; exists {
			if _, toDelete := schemaMap[key]; toDelete {
				delete(p.replicators[key], rep.Info.ID)
			} else {
				rep.Schemas = append(rep.Schemas, key)
			}
		}
	}

	if len(rep.Schemas) == 0 {
		// Remove the destination's peer multiaddress in the peerstore.
		p.host.Peerstore().ClearAddrs(rep.Info.ID)
	}

	// persist the replicator to the store, deleting it if no schemas remain
	key := core.NewReplicatorKey(rep.Info.ID.String())
	if len(rep.Schemas) == 0 {
		return txn.Systemstore().Delete(ctx, key.ToDS())
	}
	repBytes, err := json.Marshal(rep)
	if err != nil {
		return err
	}
	return txn.Systemstore().Put(ctx, key.ToDS(), repBytes)
}

func (p *Peer) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	txn, err := p.db.NewTxn(ctx, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	// create collection system prefix query
	query := dsq.Query{
		Prefix: core.NewReplicatorKey("").ToString(),
	}
	results, err := txn.Systemstore().Query(ctx, query)
	if err != nil {
		return nil, err
	}

	var reps []client.Replicator
	for result := range results.Next() {
		var rep client.Replicator
		if err = json.Unmarshal(result.Value, &rep); err != nil {
			return nil, err
		}
		reps = append(reps, rep)
	}
	return reps, nil
}
