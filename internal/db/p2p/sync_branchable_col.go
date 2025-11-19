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

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	dbid "github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// syncBranchableCollectionTopic is the fixed topic for branchable collection sync operations.
const syncBranchableCollectionTopic = "sync-branchable"

// syncBranchableCollectionRequest represents a request to synchronize a branchable collection.
type syncBranchableCollectionRequest struct {
	CollectionID string `json:"collectionID"`
}

// syncBranchableCollectionReply represents the response to a collection sync request.
type syncBranchableCollectionReply struct {
	CollectionID string `json:"collectionID"`
	Head         []byte `json:"head"` // CID bytes of the collection head
	Sender       string `json:"sender"`
}

// SyncBranchableCollection initiates a request for the latest version of the branchable
// collection's DAG from the network.
//
// This function call will block until there is a response for the collection.
// It is the responsibility of the caller to set an appropriate timeout on the context.
func (p *P2P) SyncBranchableCollection(ctx context.Context, collectionID string) error {
	cols, err := p.db.GetCollections(
		ctx,
		client.CollectionFetchOptions{
			CollectionID: immutable.Some(collectionID),
		},
	)
	if err != nil {
		return err
	}
	if len(cols) == 0 {
		return client.NewErrCollectionNotFoundForCollectionVersion(collectionID)
	}

	col := cols[0].Version()
	if !col.IsBranchable {
		return errors.New("collection is not branchable", errors.NewKV("CollectionID", collectionID))
	}

	return p.syncBranchableCollection(ctx, collectionID)
}

// syncBranchableCollection requests branchable collection synchronization from the network.
func (p *P2P) syncBranchableCollection(
	ctx context.Context,
	collectionID string,
) error {
	pubsubReq := &syncBranchableCollectionRequest{CollectionID: collectionID}

	data, err := cbor.Marshal(pubsubReq)
	if err != nil {
		return err
	}

	pubSubRespChan, err := p.host.PublishToTopic(ctx, syncBranchableCollectionTopic, data, true)
	if err != nil {
		return err
	}

	return p.waitAndHandleSyncBranchableCollectionResponse(ctx, collectionID, pubSubRespChan)
}

// waitAndHandleSyncBranchableCollectionResponse handles responses from multiple peers.
func (p *P2P) waitAndHandleSyncBranchableCollectionResponse(
	ctx context.Context,
	collectionID string,
	pubSubRespChan <-chan client.PubsubResponse,
) error {
	syncedHeads := make(map[string]cid.Cid)

loop:
	for {
		select {
		case resp := <-pubSubRespChan:
			err := p.handleSyncBranchableCollectionResponse(ctx, resp, collectionID, syncedHeads)
			if err != nil {
				return err
			}

		case <-ctx.Done():
			if len(syncedHeads) == 0 {
				return ErrTimeoutCollectionSync
			}
			break loop
		}
	}

	return nil
}

// handleSyncBranchableCollectionResponse processes a single response from a peer.
// It mutates the syncedHeads map to track which heads have been synced.
func (p *P2P) handleSyncBranchableCollectionResponse(
	ctx context.Context,
	resp client.PubsubResponse,
	collectionID string,
	syncedHeads map[string]cid.Cid,
) error {
	if resp.Err != nil {
		log.ErrorE("Received error response from peer", resp.Err)
		return resp.Err
	}

	var reply syncBranchableCollectionReply
	if err := cbor.Unmarshal(resp.Data, &reply); err != nil {
		log.ErrorE("Failed to unmarshal collection sync reply", err)
		return err
	}

	if reply.CollectionID != collectionID {
		log.ErrorE("Received response for different collection",
			errors.New("collection ID mismatch",
				errors.NewKV("Expected", collectionID),
				errors.NewKV("Received", reply.CollectionID)))
		return nil
	}

	if len(reply.Head) == 0 {
		// Peer has no commits for this collection, not an error
		return nil
	}

	_, colCid, err := cid.CidFromBytes(reply.Head)
	if err != nil {
		log.ErrorE("Failed to parse CID from reply", err)
		return err
	}

	cidStr := colCid.String()
	if _, exists := syncedHeads[cidStr]; exists {
		return nil
	}

	err = p.syncCollectionAndMerge(ctx, reply.Sender, collectionID, colCid)
	if err != nil {
		log.ErrorE("Failed to sync collection and merge", err,
			corelog.String("CollectionID", collectionID),
			corelog.String("Head", cidStr))
		return err
	}

	syncedHeads[cidStr] = colCid
	return nil
}

// syncCollectionAndMerge synchronizes a branchable collection from a remote peer and publishes a merge event.
func (p *P2P) syncCollectionAndMerge(
	ctx context.Context,
	senderID string,
	collectionID string,
	head cid.Cid,
) error {
	err := p.syncCollectionDAG(ctx, head)
	if err != nil {
		return err
	}

	evt := event.Merge{
		ByPeer:       senderID,
		FromPeer:     p.host.ID(),
		Cid:          head,
		CollectionID: collectionID,
	}

	return p.db.Merge(ctx, evt)
}

// syncCollectionDAG synchronizes the DAG for a specific branchable collection CID.
func (p *P2P) syncCollectionDAG(ctx context.Context, colCid cid.Cid) error {
	linkSys := makeLinkSystem(p.host.IPLDStore())

	nd, err := linkSys.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: colCid}, coreblock.BlockSchemaPrototype)
	if err != nil {
		return err
	}

	linkBlock, err := coreblock.GetFromNode(nd)
	if err != nil {
		return err
	}

	return p.syncDAG(ctx, linkBlock)
}

// syncBranchableCollectionMessageHandler handles incoming branchable collection sync requests from the pubsub network.
func (p *P2P) syncBranchableCollectionMessageHandler(from string, topic string, msg []byte) ([]byte, error) {
	req := &syncBranchableCollectionRequest{}
	if err := cbor.Unmarshal(msg, req); err != nil {
		return nil, err
	}

	head, err := p.processSyncBranchableCollection(req.CollectionID)
	if err != nil {
		head = []byte{}
	}

	reply := &syncBranchableCollectionReply{
		Sender:       p.host.ID(),
		CollectionID: req.CollectionID,
		Head:         head,
	}

	return cbor.Marshal(reply)
}

// processSyncBranchableCollection processes a branchable collection sync request and returns the head CID.
func (p *P2P) processSyncBranchableCollection(collectionID string) ([]byte, error) {
	clientTxn, err := p.db.NewTxn(true)
	if err != nil {
		return nil, err
	}
	defer clientTxn.Discard()

	cols, err := p.db.GetCollections(
		p.ctx,
		client.CollectionFetchOptions{
			CollectionID: immutable.Some(collectionID),
		},
	)
	if err != nil || len(cols) == 0 {
		return nil, err
	}

	// TODO: Can we add test where on node has 2 heads and another one sync with it?
	col := cols[0].Version()
	if !col.IsBranchable {
		return nil, errors.New("collection is not branchable", errors.NewKV("CollectionID", collectionID))
	}

	txn := datastore.MustGetFromClientTxn(clientTxn)

	txnCtx := dbid.InitCollectionShortIDCache(p.ctx)
	txnCtx = datastore.CtxSetTxn(txnCtx, txn)
	shortID, err := dbid.GetShortCollectionID(txnCtx, col.CollectionID)
	if err != nil {
		return nil, err
	}

	key := keys.NewHeadstoreColKey(shortID)
	headset := coreblock.NewHeadSet(txn.Headstore(), key)

	cids, _, err := headset.List(txnCtx)
	if err != nil {
		return nil, err
	}

	if len(cids) == 0 {
		return nil, errors.New("no heads found for branchable collection", errors.NewKV("CollectionID", collectionID))
	}

	return cids[0].Bytes(), nil
}
