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
	"time"

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
	CollectionID string   `json:"collectionID"`
	Heads        [][]byte `json:"heads"`
	Sender       string   `json:"sender"`
}

// SyncBranchableCollection initiates a request for the latest version of the branchable
// collection's DAG from the network.
//
// This function call will block until there is a response for the collection.
// It is the responsibility of the caller to set an appropriate timeout on the context.
func (p *P2P) SyncBranchableCollection(ctx context.Context, collectionID string) error {
	log.InfoContext(ctx, "Starting branchable collection sync",
		corelog.String("PeerID", p.host.ID()))

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
		return NewErrCollectionNotBranchable(collectionID)
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

	log.InfoContext(ctx, "Publishing sync request to topic",
		corelog.String("Topic", syncBranchableCollectionTopic))

	pubSubRespChan, err := p.host.PublishToTopic(ctx, syncBranchableCollectionTopic, data, true)
	if err != nil {
		return err
	}

	waitCtx := ctx
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	return p.waitAndHandleSyncBranchableCollectionResponse(waitCtx, collectionID, pubSubRespChan)
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
				log.InfoContext(ctx, "Sync timeout with no heads synced")
				return ErrTimeoutCollectionSync
			}
			log.InfoContext(ctx, "Sync completed (context done)",
				corelog.Int("SyncedHeadsCount", len(syncedHeads)))
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

	log.InfoContext(ctx, "Received sync response from peer",
		corelog.String("ReceiverID", p.host.ID()),
		corelog.String("Sender", reply.Sender),
		corelog.Int("HeadsCount", len(reply.Heads)))

	if reply.CollectionID != collectionID {
		log.ErrorE("Received response for different collection",
			errors.New("collection ID mismatch",
				errors.NewKV("Expected", collectionID),
				errors.NewKV("Received", reply.CollectionID)))
		return nil
	}

	if len(reply.Heads) == 0 {
		log.InfoContext(ctx, "Peer has no heads for collection",
			corelog.String("Sender", reply.Sender))
		return nil
	}

	for _, headBytes := range reply.Heads {
		_, colCid, err := cid.CidFromBytes(headBytes)
		if err != nil {
			log.ErrorE("Failed to parse CID from reply", err)
			return err
		}

		cidStr := colCid.String()
		if _, exists := syncedHeads[cidStr]; exists {
			log.InfoContext(ctx, "Skipping already synced head",
				corelog.String("CID", cidStr))
			continue
		}

		log.InfoContext(ctx, "Syncing head from peer",
			corelog.String("Sender", reply.Sender),
			corelog.String("CID", cidStr))

		err = p.syncCollectionAndMerge(ctx, reply.Sender, collectionID, colCid)
		if err != nil {
			log.ErrorE("Failed to sync collection and merge", err,
				corelog.String("Head", cidStr))
			return err
		}

		log.InfoContext(ctx, "Successfully synced and merged head",
			corelog.String("CID", cidStr))

		syncedHeads[cidStr] = colCid
	}

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

	log.Info("Received sync request from peer",
		corelog.String("From", from),
		corelog.String("PeerID", p.host.ID()))

	heads, err := p.processSyncBranchableCollection(req.CollectionID)
	if err != nil {
		log.Info("No heads to send (error retrieving)",
			corelog.Any("Error", err))
		heads = [][]byte{}
	}

	// Convert heads to CID strings for logging
	headCIDs := make([]string, len(heads))
	for i, h := range heads {
		_, c, err := cid.CidFromBytes(h)
		if err == nil {
			headCIDs[i] = c.String()
		}
	}

	log.Info("Sending sync response",
		corelog.String("SenderID", p.host.ID()),
		corelog.String("To", from),
		corelog.Int("HeadsCount", len(heads)),
		corelog.Any("HeadCIDs", headCIDs))

	reply := &syncBranchableCollectionReply{
		Sender:       p.host.ID(),
		CollectionID: req.CollectionID,
		Heads:        heads,
	}

	return cbor.Marshal(reply)
}

// processSyncBranchableCollection processes a branchable collection sync request and returns all head CIDs.
func (p *P2P) processSyncBranchableCollection(collectionID string) ([][]byte, error) {
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

	col := cols[0].Version()
	if !col.IsBranchable {
		return nil, NewErrCollectionNotBranchable(collectionID)
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
		return nil, NewErrNoHeadsForBranchableCol(collectionID)
	}

	heads := make([][]byte, len(cids))
	for i, c := range cids {
		heads[i] = c.Bytes()
	}

	return heads, nil
}
