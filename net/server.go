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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
	cid "github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	libpeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/sourcenetwork/corelog"
	rpc "github.com/sourcenetwork/go-libp2p-pubsub-rpc"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/permission"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/net/protocol"
)

const networkRequestTimeout = 10 * time.Second

// DocSyncTopic is the fixed topic for document sync operations.
const docSyncTopic = "doc-sync"

// server is the request/response instance for all P2P RPC communication.
// Implements gRPC server. See net/pb/net.proto for corresponding service definitions.
//
// Specifically, server handles the push/get request/response aspects of the RPC service
// but not the API calls.
type server struct {
	*protocol.IdentityProtocol
	*protocol.ReplicatorProtocol
	peer *Peer

	topics map[string]pubsubTopic
	// replicators is a map from collection CollectionID => peerId
	replicators map[string]map[libpeer.ID]struct{}
	mu          sync.Mutex

	docSyncTopic pubsubTopic

	conns  map[libpeer.ID]network.Stream
	connMu sync.RWMutex

	peerIdentities map[libpeer.ID]identity.Identity
	piMux          sync.RWMutex
}

// pubsubTopic is a wrapper of rpc.Topic to be able to track if the topic has
// been subscribed to.
type pubsubTopic struct {
	*rpc.Topic
	subscribed bool
}

// newServer creates a new network server that handle/directs RPC requests to the
// underlying DB instance.
func newServer(p *Peer) (*server, error) {
	s := &server{
		peer:             p,
		conns:            make(map[libpeer.ID]network.Stream),
		topics:           make(map[string]pubsubTopic),
		replicators:      make(map[string]map[libpeer.ID]struct{}),
		peerIdentities:   make(map[libpeer.ID]identity.Identity),
		IdentityProtocol: protocol.NewIdentityProtocol(p.host, p.db.GetNodeIdentityToken),
	}
	s.ReplicatorProtocol = protocol.NewReplicatorProtocol(p.host, s.processPushlog, p.handleReplicatorFailure)
	docSyncTopic, err := s.addPubSubTopic(docSyncTopic, true, s.docSyncMessageHandler)
	if err != nil {
		return nil, err
	}

	s.docSyncTopic = docSyncTopic

	return s, nil
}

// processPushlog processes a push log request
func (s *server) processPushlog(
	ctx context.Context,
	req *protocol.PushLogRequest,
	isReplicator bool,
) (*protocol.PushLogReply, error) {
	pid, err := libpeer.Decode(req.SenderID)
	if err != nil {
		return nil, errors.Wrap("parsing stream PeerID", err)
	}
	headCID, err := cid.Cast(req.CID)
	if err != nil {
		return nil, err
	}

	if req.DocID != "" {
		_, err := client.NewDocIDFromString(req.DocID)
		if err != nil {
			return nil, err
		}
	}
	byPeer, err := libpeer.Decode(req.Creator)
	if err != nil {
		return nil, err
	}
	block, err := coreblock.GetFromBytes(req.Block)
	if err != nil {
		return nil, err
	}

	// No need to check access if the message is for replication as the node sending
	// will have done so deliberately.
	if !isReplicator {
		mightHaveAccess, err := s.trySelfHasAccess(block, req.CollectionID)
		if err != nil {
			return nil, err
		}
		if !mightHaveAccess {
			// If we know we don't have access, we can skip the rest of the processing.
			return &protocol.PushLogReply{}, nil
		}
	}

	err = syncDAG(ctx, s.peer.blockService, block)
	if err != nil {
		return nil, err
	}

	s.peer.bus.Publish(event.NewMessage(event.MergeName, event.Merge{
		DocID:        req.DocID,
		ByPeer:       byPeer,
		FromPeer:     pid,
		Cid:          headCID,
		CollectionID: req.CollectionID,
	}))

	return &protocol.PushLogReply{}, nil
}

// addPubSubTopic subscribes to a topic on the pubsub network
// A custom message handler can be provided to handle incoming messages. If not provided,
// the default message handler will be used.
func (s *server) addPubSubTopic(topic string, subscribe bool, handler rpc.MessageHandler) (pubsubTopic, error) {
	if s.peer.ps == nil {
		return pubsubTopic{}, nil
	}

	log.InfoContext(s.peer.ctx, "Adding pubsub topic",
		corelog.String("PeerID", s.peer.PeerID().String()),
		corelog.String("Topic", topic))

	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.topics[topic]; ok {
		// When the topic was previously set to publish only and we now want to subscribe,
		// we need to close the existing topic and create a new one.
		if !t.subscribed && subscribe {
			if err := t.Close(); err != nil {
				return pubsubTopic{}, err
			}
		} else {
			return t, nil
		}
	}

	t, err := rpc.NewTopic(s.peer.ctx, s.peer.ps, s.peer.host.ID(), topic, subscribe)
	if err != nil {
		return pubsubTopic{}, err
	}

	if handler == nil {
		handler = s.pubSubMessageHandler
	}

	t.SetEventHandler(s.pubSubEventHandler)
	t.SetMessageHandler(handler)
	pst := pubsubTopic{
		Topic:      t,
		subscribed: subscribe,
	}
	s.topics[topic] = pst
	return pst, nil
}

func (s *server) AddPubSubTopic(topicName string, handler rpc.MessageHandler) error {
	_, err := s.addPubSubTopic(topicName, true, handler)
	return err
}

// removePubSubTopic unsubscribes to a topic
func (s *server) removePubSubTopic(topic string) error {
	if s.peer.ps == nil {
		return nil
	}

	log.Info("Removing pubsub topic",
		corelog.String("PeerID", s.peer.PeerID().String()),
		corelog.String("Topic", topic))

	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.topics[topic]; ok {
		delete(s.topics, topic)
		return t.Close()
	}
	return nil
}

func (s *server) removeAllPubsubTopics() error {
	if s.peer.ps == nil {
		return nil
	}

	log.Info("Removing all pubsub topics",
		corelog.String("PeerID", s.peer.PeerID().String()))

	s.mu.Lock()
	defer s.mu.Unlock()
	for id, t := range s.topics {
		delete(s.topics, id)
		if err := t.Close(); err != nil {
			return err
		}
	}
	return nil
}

// publishLog publishes the given PushLogRequest object on the PubSub network via the
// corresponding topic
func (s *server) publishLog(ctx context.Context, topic string, req *protocol.PushLogRequest) error {
	if s.peer.ps == nil { // skip if we aren't running with a pubsub net
		return nil
	}

	data, err := cbor.Marshal(req)
	if err != nil {
		return errors.Wrap("failed to marshal pubsub message", err)
	}

	log.InfoContext(ctx, "Publish log",
		corelog.String("PeerID", s.peer.PeerID().String()),
		corelog.String("Topic", topic))

	s.mu.Lock()
	t, ok := s.topics[topic]
	s.mu.Unlock()
	if ok {
		_, err = t.Publish(ctx, data, rpc.WithIgnoreResponse(true))
		if err != nil {
			return NewErrPushLog(err, errors.NewKV("Topic", topic))
		}
		return nil
	}

	// If the topic hasn't been explicitly subscribed to, we temporarily join it
	// to publish the log.
	return s.publishDirectToTopic(ctx, topic, data, false)
}

// publishDirectToTopic temporarily joins a pubsub topic to publish data and immediately closes it.
//
// This is useful to publish messages without incurring the cost of a full pubsub rpc topic.
func (s *server) publishDirectToTopic(ctx context.Context, topic string, data []byte, isRetry bool) error {
	psTopic, err := s.peer.ps.Join(topic)
	if err != nil {
		if strings.Contains(err.Error(), "topic already exists") && !isRetry {
			// Reaching this is really rare and probably only possible
			// through from tests. We can handle this by simply trying again a single time.
			return s.publishDirectToTopic(ctx, topic, data, true)
		}
		return NewErrPushLog(err, errors.NewKV("Topic", topic))
	}
	err = psTopic.Publish(ctx, data)
	if err != nil {
		return NewErrPushLog(err, errors.NewKV("Topic", topic))
	}
	return psTopic.Close()
}

// pubSubMessageHandler handles incoming PushLog messages from the pubsub network.
func (s *server) pubSubMessageHandler(from libpeer.ID, topic string, msg []byte) ([]byte, error) {
	log.Info("Received new pubsub event",
		corelog.String("PeerID", s.peer.PeerID().String()),
		corelog.Any("SenderId", from),
		corelog.String("Topic", topic))

	req := &protocol.PushLogRequest{}
	if err := cbor.Unmarshal(msg, req); err != nil {
		log.ErrorE("Failed to unmarshal pubsub message %s", err)
		return nil, err
	}
	req.SenderID = from.String()
	if _, err := s.processPushlog(s.peer.ctx, req, false); err != nil {
		return nil, errors.Wrap(fmt.Sprintf("Failed pushing log for doc %s", topic), err)
	}
	return nil, nil
}

// pubSubEventHandler logs events from the subscribed DocID topics.
func (s *server) pubSubEventHandler(from libpeer.ID, topic string, msg []byte) {
	log.Info("Received new pubsub event",
		corelog.String("PeerID", s.peer.PeerID().String()),
		corelog.Any("SenderId", from),
		corelog.String("Topic", topic),
		corelog.String("Message", string(msg)),
	)
	evt := event.NewMessage(event.PubSubName, event.PubSub{
		Peer: from,
	})
	s.peer.bus.Publish(evt)
}

func (s *server) updateReplicators(rep peer.AddrInfo, collectionIDs map[string]struct{}) {
	if len(collectionIDs) == 0 {
		// remove peer from store
		s.peer.host.Peerstore().ClearAddrs(rep.ID)
	} else {
		// add peer to store
		s.peer.host.Peerstore().AddAddrs(rep.ID, rep.Addrs, peerstore.PermanentAddrTTL)
		// connect to the peer
		if err := s.peer.Connect(s.peer.ctx, rep); err != nil {
			log.ErrorE("Failed to connect to replicator peer", err)
		}
	}

	// update the cached replicators
	s.mu.Lock()
	for collectionID, peers := range s.replicators {
		if _, hasID := collectionIDs[collectionID]; hasID {
			s.replicators[collectionID][rep.ID] = struct{}{}
			delete(collectionIDs, collectionID)
		} else {
			if _, exists := peers[rep.ID]; exists {
				delete(s.replicators[collectionID], rep.ID)
			}
		}
	}
	for collectionID := range collectionIDs {
		if _, exists := s.replicators[collectionID]; !exists {
			s.replicators[collectionID] = make(map[libpeer.ID]struct{})
		}
		s.replicators[collectionID][rep.ID] = struct{}{}
	}
	s.mu.Unlock()
}

func (s *server) SendPubSubMessage(
	ctx context.Context,
	topic string,
	data []byte,
) (<-chan rpc.Response, error) {
	s.mu.Lock()
	t, ok := s.topics[topic]
	s.mu.Unlock()
	if !ok {
		return nil, NewErrTopicDoesNotExist(topic)
	}
	return t.Publish(ctx, data)
}

// hasAccess checks if the requesting peer has access to the given cid.
//
// This is used as a filter in bitswap to determine if we should send the block to the requesting peer.
func (s *server) hasAccess(p libpeer.ID, c cid.Cid) bool {
	if !s.peer.documentACP.HasValue() {
		return true
	}

	clientTxn, err := s.peer.db.NewTxn(s.peer.ctx, false)
	if err != nil {
		log.ErrorE("Failed to get new transaction", err)
		return false
	}
	defer clientTxn.Discard(s.peer.ctx)
	txn := datastore.MustGetFromClientTxn(clientTxn)

	rawblock, err := txn.Blockstore().Get(s.peer.ctx, c)
	if err != nil {
		log.ErrorE("Failed to get block", err)
		return false
	}

	_, err = coreblock.GetSignatureBlockFromBytes(rawblock.RawData())
	if err == nil {
		// If the block is a signature block, we can safely send it to the requesting peer.
		return true
	}

	block, err := coreblock.GetFromBytes(rawblock.RawData())
	if err != nil {
		log.ErrorE("Failed to get doc from block", err)
		return false
	}

	cols, err := clientTxn.GetCollections(
		s.peer.ctx,
		client.CollectionFetchOptions{
			VersionID: immutable.Some(block.Delta.GetSchemaVersionID()),
		},
	)
	if err != nil {
		log.ErrorE("Failed to get collections", err)
		return false
	}
	if len(cols) == 0 {
		log.Info("No collections found", corelog.Any("Schema Version ID", block.Delta.GetSchemaVersionID()))
		return false
	}

	// If the requesting peer is in the replicators list for that collection, then they have access.
	s.mu.Lock()
	if peerList, ok := s.replicators[cols[0].SchemaRoot()]; ok {
		_, exists := peerList[p]
		if exists {
			s.mu.Unlock()
			return true
		}
	}
	s.mu.Unlock()

	identFunc := func() immutable.Option[identity.Identity] {
		s.piMux.RLock()
		ident, ok := s.peerIdentities[p]
		s.piMux.RUnlock()
		if !ok {
			ctx, cancel := context.WithTimeout(s.peer.ctx, networkRequestTimeout)
			defer cancel()
			resp, err := s.GetIdentity(ctx, p)
			if err != nil {
				log.ErrorE("Failed to get identity", err)
				return immutable.None[identity.Identity]()
			}
			ident, err = identity.FromToken(resp.IdentityToken)
			if err != nil {
				log.ErrorE("Failed to parse identity token", err)
				return immutable.None[identity.Identity]()
			}
			tokenIdent, ok := ident.(identity.TokenIdentity)
			if !ok {
				log.ErrorE("Identity is not of type TokenIdentity", nil, corelog.String("Actual", fmt.Sprintf("%T", ident)))
				return immutable.None[identity.Identity]()
			}
			err = identity.VerifyAuthToken(tokenIdent, s.peer.PeerID().String())
			if err != nil {
				log.ErrorE("Failed to verify auth token", err)
				return immutable.None[identity.Identity]()
			}
			s.piMux.Lock()
			s.peerIdentities[p] = ident
			s.piMux.Unlock()
		}
		return immutable.Some(ident)
	}

	peerHasAccess, err := permission.CheckDocAccessWithIdentityFunc(
		s.peer.ctx,
		identFunc,
		s.peer.documentACP.Value(),
		cols[0], // For now we assume there is only one collection.
		acpTypes.DocumentReadPerm,
		string(block.Delta.GetDocID()),
	)
	if err != nil {
		log.ErrorE("Failed to check access", err)
		return false
	}

	return peerHasAccess
}

// trySelfHasAccess checks if the local node has access to the given block.
//
// This is a best-effort check and returns true unless we explicitly find that the local node
// doesn't have access or if we get an error. The node sending is ultimately responsible for
// ensuring that the recipient has access.
func (s *server) trySelfHasAccess(block *coreblock.Block, p2pID string) (bool, error) {
	if !s.peer.documentACP.HasValue() {
		return true, nil
	}

	clientTxn, err := s.peer.db.NewTxn(s.peer.ctx, false)
	if err != nil {
		return false, err
	}
	defer clientTxn.Discard(s.peer.ctx)

	cols, err := clientTxn.GetCollections(
		s.peer.ctx,
		client.CollectionFetchOptions{
			CollectionID: immutable.Some(p2pID),
		},
	)
	if err != nil {
		return false, err
	}
	if len(cols) == 0 {
		return false, client.ErrCollectionNotFound
	}
	ident, err := clientTxn.GetNodeIdentity(s.peer.ctx)
	if err != nil {
		return false, err
	}
	if !ident.HasValue() {
		return true, nil
	}

	peerHasAccess, err := permission.CheckDocAccessWithIdentityFunc(
		s.peer.ctx,
		func() immutable.Option[identity.Identity] {
			return immutable.Some(identity.FromDID(ident.Value().DID))
		},
		s.peer.documentACP.Value(),
		cols[0], // For now we assume there is only one collection.
		acpTypes.DocumentReadPerm,
		string(block.Delta.GetDocID()),
	)
	if err != nil {
		return false, err
	}

	return peerHasAccess, nil
}

// docSyncMessageHandler handles incoming document sync requests from the pubsub network.
func (s *server) docSyncMessageHandler(from libpeer.ID, topic string, msg []byte) ([]byte, error) {
	req := &docSyncRequest{}
	if err := cbor.Unmarshal(msg, req); err != nil {
		return nil, err
	}

	var results []docSyncItem

	for _, docID := range req.DocIDs {
		result, err := s.processDocSyncItem(docID)
		if err != nil {
			log.ErrorE("Failed to process doc sync item", err, corelog.String("DocID", docID))
			continue // Skip failed items
		}
		results = append(results, result)
	}

	reply := &docSyncReply{
		Sender:  s.peer.host.ID().String(),
		Results: results,
	}

	return cbor.Marshal(reply)
}

// processDocSyncItem processes a single document sync request and returns the result.
func (s *server) processDocSyncItem(docID string) (docSyncItem, error) {
	txn, err := s.peer.db.NewTxn(s.peer.ctx, true)
	if err != nil {
		return docSyncItem{}, NewErrFailedToCreateTransaction(err)
	}
	defer txn.Discard(s.peer.ctx)

	key := keys.HeadstoreDocKey{
		DocID:   docID,
		FieldID: core.COMPOSITE_NAMESPACE,
	}

	headstore := datastore.HeadstoreFrom(s.peer.db.Rootstore())
	headset := coreblock.NewHeadSet(headstore, key)

	cids, _, err := headset.List(s.peer.ctx)
	if err != nil {
		return docSyncItem{}, fmt.Errorf("failed to get list of heads docID %s: %w", key.ToString(), err)
	}

	if len(cids) == 0 {
		return docSyncItem{}, fmt.Errorf("heads not found for %s", key.ToString())
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
