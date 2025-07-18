// Copyright 2024 Democratized Data Foundation
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
	"slices"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/boxo/blockservice"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/storage/bsrvadapter"
	libpeer "github.com/libp2p/go-libp2p/core/peer"

	"github.com/sourcenetwork/corelog"
	rpc "github.com/sourcenetwork/go-libp2p-pubsub-rpc"

	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
)

// syncBlockLinkTimeout is the maximum amount of time
// to wait for a block link to be fetched.
var syncBlockLinkTimeout = 5 * time.Second

func makeLinkSystem(blockService blockservice.BlockService) linking.LinkSystem {
	blockStore := &bsrvadapter.Adapter{Wrapped: blockService}

	linkSys := cidlink.DefaultLinkSystem()
	linkSys.SetWriteStorage(blockStore)
	linkSys.SetReadStorage(blockStore)
	linkSys.TrustedStorage = true

	return linkSys
}

// syncDAG synchronizes the DAG starting with the given block
// using the blockservice to fetch remote blocks.
//
// This process walks the entire DAG until the issue below is resolved.
// https://github.com/sourcenetwork/defradb/issues/2722
func syncDAG(ctx context.Context, blockService blockservice.BlockService, block *coreblock.Block) error {
	// use a session to make remote fetches more efficient
	ctx = blockservice.ContextWithSession(ctx, blockService)

	linkSystem := makeLinkSystem(blockService)

	// Store the block in the DAG store
	_, err := linkSystem.Store(linking.LinkContext{Ctx: ctx}, coreblock.GetLinkPrototype(), block.GenerateNode())
	if err != nil {
		return err
	}

	err = loadBlockLinks(ctx, &linkSystem, block)
	if err != nil {
		return err
	}
	return nil
}

// loadBlockLinks loads the links of a block recursively.
//
// If it encounters errors in the concurrent loading of links, it will return
// the first error it encountered.
func loadBlockLinks(ctx context.Context, linkSys *linking.LinkSystem, block *coreblock.Block) error {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()
	var wg sync.WaitGroup
	var asyncErr error
	var asyncErrOnce sync.Once

	// TODO: this part is not tested yet because there is not easy way of doing it at the moment.
	// https://github.com/sourcenetwork/defradb/issues/3525
	if block.Signature != nil {
		// we deliberately ignore the first returned value, which indicates whether the signature
		// the block was actually verified or not, because we don't handle it any different here.
		// But we want to keep the API of VerifyBlockSignature explicit about the results.
		_, err := coreblock.VerifyBlockSignature(block, linkSys)
		if err != nil {
			return err
		}
	}

	setAsyncErr := func(err error) {
		asyncErr = err
		cancel()
	}

	for _, lnk := range block.AllLinks() {
		wg.Add(1)
		go func(lnk cidlink.Link) {
			defer wg.Done()
			if ctxWithCancel.Err() != nil {
				return
			}
			ctxWithTimeout, cancel := context.WithTimeout(ctx, syncBlockLinkTimeout)
			defer cancel()
			nd, err := linkSys.Load(linking.LinkContext{Ctx: ctxWithTimeout}, lnk, coreblock.BlockSchemaPrototype)
			if err != nil {
				asyncErrOnce.Do(func() { setAsyncErr(err) })
				return
			}
			linkBlock, err := coreblock.GetFromNode(nd)
			if err != nil {
				asyncErrOnce.Do(func() { setAsyncErr(err) })
				return
			}

			err = loadBlockLinks(ctx, linkSys, linkBlock)
			if err != nil {
				asyncErrOnce.Do(func() { setAsyncErr(err) })
				return
			}
		}(lnk)
	}

	wg.Wait()

	return asyncErr
}

// syncDocuments requests document synchronization from the network.
func (s *server) syncDocuments(
	ctx context.Context,
	collectionID string,
	docIDs []string,
) (map[string][]cid.Cid, error) {
	pubsubReq := &docSyncRequest{DocIDs: docIDs}

	data, err := cbor.Marshal(pubsubReq)
	if err != nil {
		return nil, err
	}

	pubSubRespChan, err := s.docSyncTopic.Publish(ctx, data, rpc.WithMultiResponse(true))
	if err != nil {
		return nil, err
	}

	return s.waitAndHandleDocSyncResponses(ctx, collectionID, docIDs, pubSubRespChan)
}

// waitAndHandleDocSyncResponses handles multiple responses from different peers.
func (s *server) waitAndHandleDocSyncResponses(
	ctx context.Context,
	collectionID string,
	docIDs []string,
	pubSubRespChan <-chan rpc.Response,
) (results map[string][]cid.Cid, err error) {
	result := make(map[string][]cid.Cid)

loop:
	for {
		select {
		case resp := <-pubSubRespChan:
			s.handleDocSyncResponse(ctx, resp, collectionID, result)

			if len(result) >= len(docIDs) {
				break loop
			}

		case <-ctx.Done():
			if len(result) == 0 {
				return nil, ErrTimeoutDocSync
			}
			break loop
		}
	}

	return result, nil
}

// handleDocSyncResponse processes a single response from a peer.
func (s *server) handleDocSyncResponse(
	ctx context.Context,
	resp rpc.Response,
	collectionID string,
	results map[string][]cid.Cid,
) {
	if resp.Err != nil {
		log.ErrorE("Received error response from peer", resp.Err)
		return
	}

	var reply docSyncReply
	if err := cbor.Unmarshal(resp.Data, &reply); err != nil {
		log.ErrorE("Failed to unmarshal doc sync reply", err)
		return
	}

	sender, err := libpeer.Decode(reply.Sender)
	if err != nil {
		log.ErrorE("Failed to decode peer id of sender", err)
		return
	}

	for _, item := range reply.Results {
		s.handleDocSyncItem(ctx, item, sender, collectionID, results)
	}
}

// handleDocSyncItem handles a single document sync item from a peer response.
func (s *server) handleDocSyncItem(
	ctx context.Context,
	item docSyncItem,
	sender libpeer.ID,
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

		err = s.syncDocumentAndMerge(ctx, sender, collectionID, item.DocID, docCid)
		if err != nil {
			log.ErrorE("Failed to sync document", err,
				corelog.String("DocID", item.DocID),
				corelog.String("CID", docCid.String()))
			continue
		}
	}
}

// syncDocumentAndMerge synchronizes a document from a remote peer and publishes a merge event.
func (s *server) syncDocumentAndMerge(
	ctx context.Context,
	sender libpeer.ID,
	collectionID, docID string,
	head cid.Cid,
) error {
	err := s.syncDocumentDAG(ctx, head)

	if err != nil {
		return err
	}

	s.peer.bus.Publish(event.NewMessage(event.MergeName, event.Merge{
		DocID:        docID,
		ByPeer:       sender,
		FromPeer:     s.peer.PeerInfo().ID,
		Cid:          head,
		CollectionID: collectionID,
	}))

	return nil
}

// syncDocumentDAG synchronizes the DAG for a specific document CID.
func (s *server) syncDocumentDAG(ctx context.Context, docCid cid.Cid) error {
	linkSys := makeLinkSystem(s.peer.blockService)

	nd, err := linkSys.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: docCid}, coreblock.BlockSchemaPrototype)
	if err != nil {
		return err
	}

	linkBlock, err := coreblock.GetFromNode(nd)
	if err != nil {
		return err
	}

	return syncDAG(ctx, s.peer.blockService, linkBlock)
}
