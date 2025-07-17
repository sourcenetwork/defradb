// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package net

import (
	"context"
	"slices"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/libp2p/go-libp2p/core/peer"
	libpeer "github.com/libp2p/go-libp2p/core/peer"

	"github.com/sourcenetwork/corelog"
	rpc "github.com/sourcenetwork/go-libp2p-pubsub-rpc"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
)

var (
	// DialTimeout is the max time duration to wait when dialing a peer.
	PushTimeout = time.Second * 10
	PullTimeout = time.Second * 10
)

// pushLog creates a pushLog request and sends it to another node
// over libp2p grpc connection
func (s *server) pushLog(evt event.Update, pid peer.ID) (err error) {
	defer func() {
		// When the event is a retry, we don't need to republish the failure as
		// it is already being handled by the retry mechanism through the success channel.
		if err != nil && !evt.IsRetry {
			handleRepErr := s.peer.handleReplicatorFailure(s.peer.ctx, pid.String(), evt.DocID)
			if handleRepErr != nil {
				err = errors.Join(err, handleRepErr)
			}
		}
	}()

	client, err := s.dial(pid) // grpc dial over P2P stream
	if err != nil {
		return NewErrPushLog(err)
	}

	ctx, cancel := context.WithTimeout(s.peer.ctx, PushTimeout)
	defer cancel()

	req := pushLogRequest{
		DocID:        evt.DocID,
		CID:          evt.Cid.Bytes(),
		CollectionID: evt.CollectionID,
		Creator:      s.peer.host.ID().String(),
		Block:        evt.Block,
	}
	if err := client.Invoke(ctx, servicePushLogName, req, nil); err != nil {
		return NewErrPushLog(
			err,
			errors.NewKV("CID", evt.Cid),
			errors.NewKV("DocID", evt.DocID),
			errors.NewKV("PeerID", pid),
		)
	}
	return nil
}

// getIdentity creates a getIdentity request and sends it to another node
func (s *server) getIdentity(ctx context.Context, pid peer.ID) (getIdentityReply, error) {
	client, err := s.dial(pid) // grpc dial over P2P stream
	if err != nil {
		return getIdentityReply{}, NewErrPushLog(err)
	}

	ctx, cancel := context.WithTimeout(ctx, PushTimeout)
	defer cancel()

	req := getIdentityRequest{
		PeerID: s.peer.host.ID().String(),
	}
	resp := getIdentityReply{}
	if err := client.Invoke(ctx, serviceGetIdentityName, req, &resp); err != nil {
		return getIdentityReply{}, NewErrFailedToGetIdentity(
			err,
			errors.NewKV("PeerID", pid),
		)
	}
	return resp, nil
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

	return s.processDocSyncResponses(ctx, collectionID, docIDs, pubSubRespChan)
}

// processDocSyncResponses handles multiple responses from different peers.
func (s *server) processDocSyncResponses(
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
			s.processDocSyncResponse(ctx, resp, collectionID, result)

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

// processDocSyncResponse processes a single response from a peer.
func (s *server) processDocSyncResponse(
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
