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
	"fmt"
	"slices"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// DocSyncTopic is the fixed topic for document sync operations.
const docSyncTopic = "doc-sync"

// docSyncRequest represents a request to synchronize specific documents.
type docSyncRequest struct {
	DocIDs []string `json:"docIDs"`
}

// docSyncReply represents the response to a document sync request.
type docSyncReply struct {
	Results []docSyncItem `json:"results"`
	Sender  string        `json:"sender"`
}

// docSyncItem represents the sync result for a single document.
type docSyncItem struct {
	DocID string   `json:"docID"`
	Heads [][]byte `json:"heads"`
}

// SyncDocuments initiates a request for the latest document versions of the
// documents corresponding to the provided docIDs list.
//
// This function call will block until there is a response for all of the docIDs listed.
// It is the responsibility of the caller to set an appropriate timeout on the context.
func (p *P2P) SyncDocuments(ctx context.Context, collectionName string, docIDs []string) error {
	log.InfoContext(ctx, "Starting document sync",
		corelog.Any("CollectionName", collectionName),
		corelog.Int("DocIDCount", len(docIDs)))

	cols, err := p.db.GetCollections(
		ctx,
		options.WithIdentity(
			options.GetCollections().SetCollectionName(collectionName),
			iIdentity.FromContext(ctx),
		),
	)
	if err != nil {
		return err
	}
	if len(cols) == 0 {
		return client.NewErrCollectionNotFoundForName(collectionName)
	}

	collectionID := cols[0].Version().CollectionID
	_, err = p.syncDocuments(ctx, collectionID, docIDs)
	if err != nil {
		return err
	}

	log.InfoContext(ctx, "Document sync completed", corelog.Any("CollectionName", collectionName))
	return nil
}

// syncDocuments requests document synchronization from the network.
func (p *P2P) syncDocuments(
	ctx context.Context,
	collectionID string,
	docIDs []string,
) (map[string][]cid.Cid, error) {
	activePeers, err := p.ActivePeers(ctx)
	if err != nil {
		return nil, err
	}

	if len(activePeers) == 0 {
		return nil, ErrTimeoutDocSync
	}

	pendingPeers := make(map[string]struct{}, len(activePeers))
	for _, peer := range activePeers {
		pendingPeers[peer] = struct{}{}
	}

	pubsubReq := &docSyncRequest{DocIDs: docIDs}

	data, err := cbor.Marshal(pubsubReq)
	if err != nil {
		return nil, err
	}

	pubSubRespChan, err := p.host.PublishToTopic(ctx, docSyncTopic, data, true)
	if err != nil {
		return nil, err
	}

	waitCtx := ctx
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	return p.waitAndHandleDocSyncResponses(waitCtx, collectionID, docIDs, pubSubRespChan, pendingPeers)
}

// waitAndHandleDocSyncResponses handles responses from multiple peers.
// It waits for all peers to respond or timeout, collecting all document heads.
func (p *P2P) waitAndHandleDocSyncResponses(
	ctx context.Context,
	collectionID string,
	docIDs []string,
	pubSubRespChan <-chan client.PubsubResponse,
	pendingPeers map[string]struct{},
) (map[string][]cid.Cid, error) {
	result := make(map[string][]cid.Cid)

	requestedDocIDs := make(map[string]struct{}, len(docIDs))
	for _, docID := range docIDs {
		requestedDocIDs[docID] = struct{}{}
	}

	for len(pendingPeers) > 0 {
		select {
		case resp := <-pubSubRespChan:
			senderID := p.handleDocSyncResponse(ctx, resp, collectionID, requestedDocIDs, result)
			delete(pendingPeers, senderID)

		case <-ctx.Done():
			if len(result) == 0 {
				return nil, ErrTimeoutDocSync
			}
			return result, nil
		}
	}

	return result, nil
}

// handleDocSyncResponse processes a single response from a peer.
// It validates docIDs against the requested set and mutates the results map.
// Returns the sender ID for peer tracking.
func (p *P2P) handleDocSyncResponse(
	ctx context.Context,
	resp client.PubsubResponse,
	collectionID string,
	requestedDocIDs map[string]struct{},
	results map[string][]cid.Cid,
) string {
	if resp.Err != nil {
		log.ErrorE("Received error response from peer", resp.Err, corelog.String("PeerID", resp.From))
		return resp.From
	}

	var reply docSyncReply
	if err := cbor.Unmarshal(resp.Data, &reply); err != nil {
		log.ErrorE("Failed to unmarshal doc sync reply", err)
		return resp.From
	}

	for _, item := range reply.Results {
		if _, ok := requestedDocIDs[item.DocID]; !ok {
			log.ErrorE("Received unrequested docID",
				errors.New("docID not in request"),
				corelog.String("DocID", item.DocID),
				corelog.String("Sender", reply.Sender))
			continue
		}
		p.handleDocSyncItem(ctx, item, reply.Sender, collectionID, results)
	}

	return reply.Sender
}

// handleDocSyncItem handles a single document sync item from a peer response.
// It mutates the results map with the document IDs and their corresponding CIDs.
func (p *P2P) handleDocSyncItem(
	ctx context.Context,
	item docSyncItem,
	senderID string,
	collectionID string,
	results map[string][]cid.Cid,
) {
	for _, headBytes := range item.Heads {
		_, docCid, err := cid.CidFromBytes(headBytes)
		if err != nil {
			log.ErrorE("Failed to parse CID from bytes", err,
				corelog.String("DocID", item.DocID))
			continue
		}

		if heads, exists := results[item.DocID]; exists {
			if !slices.Contains(heads, docCid) {
				results[item.DocID] = append(heads, docCid)
			} else {
				// we've seen this head already, just skip
				continue
			}
		} else {
			results[item.DocID] = []cid.Cid{docCid}
		}

		err = p.syncDocumentAndMerge(ctx, senderID, collectionID, item.DocID, docCid)
		if err != nil {
			log.ErrorE("Failed to sync document", err,
				corelog.String("DocID", item.DocID),
				corelog.String("CID", docCid.String()))
			continue
		}
	}
}

// syncDocumentAndMerge synchronizes a document from a remote peer and publishes a merge event.
func (p *P2P) syncDocumentAndMerge(
	ctx context.Context,
	senderID string,
	collectionID, docID string,
	head cid.Cid,
) error {
	err := p.syncDocumentDAG(ctx, head)
	if err != nil {
		return err
	}

	evt := event.Merge{
		DocID:        docID,
		ByPeer:       senderID,
		FromPeer:     p.host.ID(),
		Cid:          head,
		CollectionID: collectionID,
	}

	return p.db.Merge(ctx, evt)
}

// syncDocumentDAG synchronizes the DAG for a specific document CID.
func (p *P2P) syncDocumentDAG(ctx context.Context, docCid cid.Cid) error {
	linkSys := makeLinkSystem(p.host.IPLDStore())

	nd, err := linkSys.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: docCid}, coreblock.BlockSchemaPrototype)
	if err != nil {
		return err
	}

	linkBlock, err := coreblock.GetFromNode(nd)
	if err != nil {
		return err
	}

	return p.syncDAG(ctx, linkBlock)
}

// docSyncMessageHandler handles incoming document sync requests from the pubsub network.
func (p *P2P) docSyncMessageHandler(from string, topic string, msg []byte) ([]byte, error) {
	req := &docSyncRequest{}
	if err := cbor.Unmarshal(msg, req); err != nil {
		return nil, err
	}

	var results []docSyncItem

	for _, docID := range req.DocIDs {
		result, err := p.processDocSyncItem(docID)
		if err != nil {
			log.ErrorE("Failed to process doc sync item", err, corelog.String("DocID", docID))
			continue
		}
		if len(result.Heads) > 0 {
			results = append(results, result)
		}
	}

	reply := &docSyncReply{
		Sender:  p.host.ID(),
		Results: results,
	}

	return cbor.Marshal(reply)
}

// processDocSyncItem processes a single document sync request and returns the result.
func (p *P2P) processDocSyncItem(docID string) (docSyncItem, error) {
	key := keys.HeadstoreDocKey{
		DocID:   docID,
		FieldID: core.COMPOSITE_NAMESPACE,
	}

	headset := coreblock.NewHeadSet(p.db.Multistore().Headstore(), key)

	cids, _, err := headset.List(p.ctx)
	if err != nil {
		return docSyncItem{}, fmt.Errorf("failed to get list of heads docID %s: %w", key.ToString(), err)
	}

	if len(cids) == 0 {
		// This node doesn't have this document — normal in a broadcast sync where
		// all subscribed nodes receive the request regardless of whether they own the doc.
		return docSyncItem{}, nil
	}

	result := docSyncItem{
		DocID: docID,
		Heads: make([][]byte, 0, len(cids)),
	}

	for _, cid := range cids {
		result.Heads = append(result.Heads, cid.Bytes())
	}

	return result, nil
}
