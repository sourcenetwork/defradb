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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	dbid "github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// collectionSyncTopic is the fixed topic for branchable collection sync operations.
const collectionSyncTopic = "collection-sync"

// collectionSyncRequest represents a request to synchronize a branchable collection.
type collectionSyncRequest struct {
	CollectionName string `json:"collectionName"`
}

// collectionSyncReply represents the response to a collection sync request.
type collectionSyncReply struct {
	CollectionName string `json:"collectionName"`
	Head           []byte `json:"head"` // CID bytes of the collection head
	Sender         string `json:"sender"`
}

// SyncBranchableCollection initiates a request for the latest version of the branchable
// collection's DAG from the network.
//
// This function call will block until there is a response for the collection.
// It is the responsibility of the caller to set an appropriate timeout on the context.
func (p *P2P) SyncBranchableCollection(ctx context.Context, collectionName string) error {
	cols, err := p.db.GetCollections(
		ctx,
		client.CollectionFetchOptions{
			Name: immutable.Some(collectionName),
		},
	)
	if err != nil {
		return err
	}
	if len(cols) == 0 {
		return client.NewErrCollectionNotFoundForName(collectionName)
	}

	col := cols[0].Version()
	if !col.IsBranchable {
		return errors.New("collection is not branchable", errors.NewKV("Name", collectionName))
	}

	collectionID := col.CollectionID
	_, err = p.syncBranchableCollection(ctx, collectionName, collectionID)
	return err
}

// syncBranchableCollection requests branchable collection synchronization from the network.
func (p *P2P) syncBranchableCollection(
	ctx context.Context,
	collectionName string,
	collectionID string,
) (cid.Cid, error) {
	pubsubReq := &collectionSyncRequest{CollectionName: collectionName}

	data, err := cbor.Marshal(pubsubReq)
	if err != nil {
		return cid.Undef, err
	}

	pubSubRespChan, err := p.host.PublishToTopic(ctx, collectionSyncTopic, data, true)
	if err != nil {
		return cid.Undef, err
	}

	return p.waitAndHandleCollectionSyncResponse(ctx, collectionID, collectionName, pubSubRespChan)
}

// waitAndHandleCollectionSyncResponse handles the response from a peer.
func (p *P2P) waitAndHandleCollectionSyncResponse(
	ctx context.Context,
	collectionID string,
	collectionName string,
	pubSubRespChan <-chan client.PubsubResponse,
) (cid.Cid, error) {
	select {
	case resp := <-pubSubRespChan:
		return p.handleCollectionSyncResponse(ctx, resp, collectionID, collectionName)

	case <-ctx.Done():
		return cid.Undef, ErrTimeoutCollectionSync
	}
}

// handleCollectionSyncResponse processes a single response from a peer.
func (p *P2P) handleCollectionSyncResponse(
	ctx context.Context,
	resp client.PubsubResponse,
	collectionID string,
	collectionName string,
) (cid.Cid, error) {
	if resp.Err != nil {
		return cid.Undef, resp.Err
	}

	var reply collectionSyncReply
	if err := cbor.Unmarshal(resp.Data, &reply); err != nil {
		return cid.Undef, err
	}

	if reply.CollectionName != collectionName {
		return cid.Undef, errors.New("received response for different collection",
			errors.NewKV("Expected", collectionName),
			errors.NewKV("Received", reply.CollectionName))
	}

	if len(reply.Head) == 0 {
		return cid.Undef, errors.New("peer has no commits for collection",
			errors.NewKV("Name", collectionName))
	}

	_, colCid, err := cid.CidFromBytes(reply.Head)
	if err != nil {
		return cid.Undef, err
	}

	err = p.syncCollectionAndMerge(ctx, reply.Sender, collectionID, collectionName, colCid)
	if err != nil {
		return cid.Undef, err
	}

	return colCid, nil
}

// syncCollectionAndMerge synchronizes a branchable collection from a remote peer and publishes a merge event.
func (p *P2P) syncCollectionAndMerge(
	ctx context.Context,
	senderID string,
	collectionID string,
	collectionName string,
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

// collectionSyncMessageHandler handles incoming branchable collection sync requests from the pubsub network.
func (p *P2P) collectionSyncMessageHandler(from string, topic string, msg []byte) ([]byte, error) {
	req := &collectionSyncRequest{}
	if err := cbor.Unmarshal(msg, req); err != nil {
		return nil, err
	}

	head, err := p.processCollectionSyncItem(req.CollectionName)
	if err != nil {
		head = []byte{}
	}

	reply := &collectionSyncReply{
		Sender:         p.host.ID(),
		CollectionName: req.CollectionName,
		Head:           head,
	}

	return cbor.Marshal(reply)
}

// processCollectionSyncItem processes a branchable collection sync request and returns the head CID.
func (p *P2P) processCollectionSyncItem(collectionName string) ([]byte, error) {
	clientTxn, err := p.db.NewTxn(true)
	if err != nil {
		return nil, err
	}
	defer clientTxn.Discard()

	cols, err := p.db.GetCollections(
		p.ctx,
		client.CollectionFetchOptions{
			Name: immutable.Some(collectionName),
		},
	)
	if err != nil || len(cols) == 0 {
		return nil, err
	}

	col := cols[0].Version()
	if !col.IsBranchable {
		return nil, errors.New("collection is not branchable", errors.NewKV("Name", collectionName))
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
		return nil, errors.New("no heads found for branchable collection", errors.NewKV("Name", collectionName))
	}

	return cids[0].Bytes(), nil
}
