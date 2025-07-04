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
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
	cid "github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/storage/bsrvadapter"
	"github.com/libp2p/go-libp2p/core/peer"
	libpeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/sourcenetwork/corelog"
	rpc "github.com/sourcenetwork/go-libp2p-pubsub-rpc"
	"github.com/sourcenetwork/immutable"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpcpeer "google.golang.org/grpc/peer"

	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/permission"
	"github.com/sourcenetwork/defradb/internal/se"
	secore "github.com/sourcenetwork/defradb/internal/se/core"
)

// server is the request/response instance for all P2P RPC communication.
// Implements gRPC server. See net/pb/net.proto for corresponding service definitions.
//
// Specifically, server handles the push/get request/response aspects of the RPC service
// but not the API calls.
type server struct {
	peer *Peer
	opts []grpc.DialOption

	topics map[string]pubsubTopic
	// replicators is a map from collection CollectionID => peerId
	replicators map[string]map[libpeer.ID]struct{}
	mu          sync.Mutex

	conns map[libpeer.ID]*grpc.ClientConn

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
func newServer(p *Peer, opts ...grpc.DialOption) (*server, error) {
	s := &server{
		peer:           p,
		conns:          make(map[libpeer.ID]*grpc.ClientConn),
		topics:         make(map[string]pubsubTopic),
		replicators:    make(map[string]map[libpeer.ID]struct{}),
		peerIdentities: make(map[libpeer.ID]identity.Identity),
	}

	cred := insecure.NewCredentials()
	defaultOpts := []grpc.DialOption{
		s.getLibp2pDialer(),
		grpc.WithTransportCredentials(cred),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype(cborCodecName)),
	}

	s.opts = append(defaultOpts, opts...)

	return s, nil
}

// pushLogHandler receives a push log request from the grpc server (replicator)
func (s *server) pushLogHandler(ctx context.Context, req *pushLogRequest) (*pushLogReply, error) {
	return s.processPushlog(ctx, req, true)
}

// processPushlog processes a push log request
func (s *server) processPushlog(
	ctx context.Context,
	req *pushLogRequest,
	isReplicator bool,
) (*pushLogReply, error) {
	pid, err := peerIDFromContext(ctx)
	if err != nil {
		return nil, err
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
			return &pushLogReply{}, nil
		}
	}

	log.InfoContext(ctx, "Received pushlog",
		corelog.Any("PeerID", pid.String()),
		corelog.Any("Creator", byPeer.String()),
		corelog.Any("DocID", req.DocID))

	err = syncDAG(ctx, s.peer.blockService, block)
	if err != nil {
		return nil, err
	}

	log.InfoContext(ctx, "DAG sync complete",
		corelog.Any("PeerID", pid.String()),
		corelog.Any("DocID", req.DocID))

	// Once processed, subscribe to the DocID topic on the pubsub network unless we already
	// subscribed to the collection.
	if !s.hasPubSubTopicAndSubscribed(req.CollectionID) && req.DocID != "" {
		_, err = s.addPubSubTopic(req.DocID, true, nil)
		if err != nil {
			return nil, err
		}
	}

	s.peer.bus.Publish(event.NewMessage(event.MergeName, event.Merge{
		DocID:        req.DocID,
		ByPeer:       byPeer,
		FromPeer:     pid,
		Cid:          headCID,
		CollectionID: req.CollectionID,
	}))

	return &pushLogReply{}, nil
}

// getIdentityHandler receives a get identity request and returns the identity token
// with the requesting peer as the audience.
func (s *server) getIdentityHandler(
	ctx context.Context,
	req *getIdentityRequest,
) (*getIdentityReply, error) {
	if !s.peer.documentACP.HasValue() {
		return &getIdentityReply{}, nil
	}
	token, err := s.peer.db.GetNodeIdentityToken(ctx, immutable.Some(req.PeerID))
	if err != nil {
		return nil, err
	}
	return &getIdentityReply{IdentityToken: token}, nil
}

// pushSEArtifactsHandler receives SE artifacts from peers
func (s *server) pushSEArtifactsHandler(ctx context.Context, req *pushSEArtifactsRequest) (*pushSEArtifactsReply, error) {
	pid, err := peerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	log.InfoContext(ctx, "Received SE artifacts",
		corelog.Any("PeerID", pid.String()),
		corelog.Any("Creator", req.Creator),
		corelog.Any("CollectionID", req.CollectionID),
		corelog.Any("ArtifactCount", len(req.Artifacts)))

	artifacts := make([]secore.Artifact, len(req.Artifacts))
	for i, netArtifact := range req.Artifacts {
		artifacts[i] = secore.Artifact{
			DocID:        netArtifact.DocID,
			IndexID:      netArtifact.IndexID,
			SearchTag:    netArtifact.SearchTag,
			CollectionID: req.CollectionID,
		}
	}

	// Store artifacts directly in the datastore
	if err := se.StoreArtifacts(ctx, datastore.DatastoreFrom(s.peer.db.Rootstore()), artifacts); err != nil {
		log.ErrorContextE(ctx, "Failed to store SE artifacts", err)
		return nil, err
	}

	return &pushSEArtifactsReply{}, nil
}

// querySEArtifactsHandler handles SE queries from peers
func (s *server) querySEArtifactsHandler(ctx context.Context, req *querySEArtifactsRequest) (*querySEArtifactsReply, error) {
	pid, err := peerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	log.InfoContext(ctx, "Received SE query",
		corelog.Any("PeerID", pid.String()),
		corelog.Any("CollectionID", req.CollectionID),
		corelog.Any("QueryCount", len(req.Queries)))

	matchingDocIDs, err := s.querySEArtifactsFromDatastore(ctx, req)
	if err != nil {
		return nil, err
	}

	return &querySEArtifactsReply{
		DocIDs: matchingDocIDs,
	}, nil
}

// querySEArtifactsFromDatastore queries SE artifacts from the local datastore
func (s *server) querySEArtifactsFromDatastore(ctx context.Context, req *querySEArtifactsRequest) ([]string, error) {
	queries := make([]se.FieldQuery, len(req.Queries))
	for i, q := range req.Queries {
		queries[i] = se.FieldQuery{
			FieldName: q.FieldName,
			IndexID:   q.IndexID,
			SearchTag: q.SearchTag,
		}
	}

	return se.FetchDocIDs(ctx, datastore.DatastoreFrom(s.peer.db.Rootstore()), req.CollectionID, queries)
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

// hasPubSubTopicAndSubscribed checks if we are subscribed to a topic.
func (s *server) hasPubSubTopicAndSubscribed(topic string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.topics[topic]
	return ok && t.subscribed
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
func (s *server) publishLog(ctx context.Context, topic string, req *pushLogRequest) error {
	if s.peer.ps == nil { // skip if we aren't running with a pubsub net
		return nil
	}
	s.mu.Lock()
	t, ok := s.topics[topic]
	s.mu.Unlock()
	if !ok {
		subscribe := topic != req.CollectionID && !s.hasPubSubTopicAndSubscribed(req.CollectionID)
		_, err := s.addPubSubTopic(topic, subscribe, nil)
		if err != nil {
			return errors.Wrap(fmt.Sprintf("failed to created single use topic %s", topic), err)
		}
		return s.publishLog(ctx, topic, req)
	}

	if topic == req.CollectionID && req.DocID == "" && !t.subscribed {
		// If the push log request is scoped to the schema and not to a document, subscribe to the
		// schema.
		var err error
		t, err = s.addPubSubTopic(topic, true, nil)
		if err != nil {
			return errors.Wrap(fmt.Sprintf("failed to created single use topic %s", topic), err)
		}
	}

	log.InfoContext(ctx, "Publish log",
		corelog.String("PeerID", s.peer.PeerID().String()),
		corelog.String("Topic", topic))

	data, err := cbor.Marshal(req)
	if err != nil {
		return errors.Wrap("failed to marshal pubsub message", err)
	}

	_, err = t.Publish(ctx, data, rpc.WithIgnoreResponse(true))
	if err != nil {
		return errors.Wrap(fmt.Sprintf("failed publishing to thread %s", topic), err)
	}
	return nil
}

// pubSubMessageHandler handles incoming PushLog messages from the pubsub network.
func (s *server) pubSubMessageHandler(from libpeer.ID, topic string, msg []byte) ([]byte, error) {
	log.Info("Received new pubsub event",
		corelog.String("PeerID", s.peer.PeerID().String()),
		corelog.Any("SenderId", from),
		corelog.String("Topic", topic))

	req := &pushLogRequest{}
	if err := cbor.Unmarshal(msg, req); err != nil {
		log.ErrorE("Failed to unmarshal pubsub message %s", err)
		return nil, err
	}
	ctx := grpcpeer.NewContext(s.peer.ctx, &grpcpeer.Peer{
		Addr: addr{from},
	})
	if _, err := s.processPushlog(ctx, req, false); err != nil {
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

// addr implements net.Addr and holds a libp2p peer ID.
type addr struct{ id libpeer.ID }

// Network returns the name of the network that this address belongs to (libp2p).
func (a addr) Network() string { return "libp2p" }

// String returns the peer ID of this address in string form (B58-encoded).
func (a addr) String() string { return a.id.String() }

// peerIDFromContext returns peer ID from the GRPC context
func peerIDFromContext(ctx context.Context) (libpeer.ID, error) {
	ctxPeer, ok := grpcpeer.FromContext(ctx)
	if !ok {
		return "", errors.New("unable to identify stream peer")
	}
	pid, err := libpeer.Decode(ctxPeer.Addr.String())
	if err != nil {
		return "", errors.Wrap("parsing stream PeerID", err)
	}
	return pid, nil
}

func (s *server) updatePubSubTopics(evt event.P2PTopic) {
	for _, topic := range evt.ToAdd {
		_, err := s.addPubSubTopic(topic, true, nil)
		if err != nil {
			log.ErrorE("Failed to add pubsub topic.", err)
		}
	}

	for _, topic := range evt.ToRemove {
		err := s.removePubSubTopic(topic)
		if err != nil {
			log.ErrorE("Failed to remove pubsub topic.", err)
		}
	}
	s.peer.bus.Publish(event.NewMessage(event.P2PTopicCompletedName, nil))
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
			resp, err := s.getIdentity(s.peer.ctx, p)
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

func (s *server) handleDocUpdateRequest(req event.DocUpdateRequest) {
	pubsubReq := &docUpdateRequest{
		CollectionID: req.CollectionID,
		DocID:        req.DocID,
		RequestorID:  s.peer.PeerID().String(),
	}

	data, err := cbor.Marshal(pubsubReq)
	if err != nil {
		req.Response <- event.DocUpdateResponse{
			Found: false,
			Error: errors.Wrap("failed to marshal doc update request", err),
		}
		return
	}

	respChan, err := s.SendPubSubMessage(s.peer.ctx, onDemandDocUpdateTopic, data)
	if err != nil {
		req.Response <- event.DocUpdateResponse{
			Found: false,
			Error: errors.Wrap("failed to publish doc update request", err),
		}
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(s.peer.ctx, 5*time.Second)
		defer cancel()

		select {
		case resp := <-respChan:
			if resp.Err != nil {
				req.Response <- event.DocUpdateResponse{
					Found: false,
					Error: resp.Err,
				}
				return
			}
			if len(resp.Data) > 0 {
				var docUpdateReply docUpdateReply
				if err := cbor.Unmarshal(resp.Data, &docUpdateReply); err != nil {
					log.ErrorContextE(ctx, "Failed to unmarshal doc update response", err)
					return
				}

				blockStore := &bsrvadapter.Adapter{Wrapped: s.peer.blockService}

				linkSys := cidlink.DefaultLinkSystem()
				linkSys.SetReadStorage(blockStore)
				linkSys.TrustedStorage = true

				_, docCid, err := cid.CidFromBytes(docUpdateReply.CID)
				if err != nil {
					log.ErrorContextE(ctx, "Failed to convert CID from bytes", err)
					return
				}

				nd, err := linkSys.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: docCid}, coreblock.BlockSchemaPrototype)
				if err != nil {
					log.ErrorContextE(ctx, "Failed to load document node", err)
					return
				}
				linkBlock, err := coreblock.GetFromNode(nd)
				if err != nil {
					log.ErrorContextE(ctx, "Failed to get block from node", err)
					return
				}

				err = syncDAG(ctx, s.peer.blockService, linkBlock)
				if err != nil {
					log.ErrorContextE(ctx, "Failed to sync DAG", err)
					return
				}

				req.Response <- event.DocUpdateResponse{Found: true}
			}
		case <-ctx.Done():
			req.Response <- event.DocUpdateResponse{
				Found: false,
				Error: err,
			}
		}
	}()
}

// docUpdateMessageHandler handles incoming document update requests from the pubsub network.
func (s *server) docUpdateMessageHandler(from libpeer.ID, topic string, msg []byte) ([]byte, error) {
	log.Info("Received doc update request",
		corelog.String("PeerID", s.peer.PeerID().String()),
		corelog.Any("SenderId", from),
		corelog.String("Topic", topic))

	req := &docUpdateRequest{}
	if err := cbor.Unmarshal(msg, req); err != nil {
		log.ErrorE("Failed to unmarshal doc update request", err)
		return nil, err
	}

	if req.RequestorID == s.peer.PeerID().String() {
		return []byte{}, nil
	}

	cols, err := s.peer.db.GetCollections(s.peer.ctx, client.CollectionFetchOptions{
		CollectionID: immutable.Some(req.CollectionID),
	})

	if err != nil {
		log.ErrorE("Failed to get collections", err)
		return []byte{}, nil
	}

	if len(cols) == 0 {
		return []byte{}, nil
	}

	col := cols[0]
	docIDStr, err := client.NewDocIDFromString(req.DocID)
	if err != nil {
		log.ErrorE("Failed to parse DocID", err)
		return []byte{}, nil
	}

	doc, err := col.Get(s.peer.ctx, docIDStr, false)
	if err != nil {
		log.ErrorE("Failed to get document", err)
		return []byte{}, nil
	}

	reply := &docUpdateReply{
		DocID:        docIDStr.String(),
		CID:          doc.Head().Bytes(),
		CollectionID: col.SchemaRoot(),
		Sender:       s.peer.host.ID().String(),
	}

	return cbor.Marshal(reply)
}
