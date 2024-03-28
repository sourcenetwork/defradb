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

	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/libp2p/go-libp2p/core/event"
	libpeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/sourcenetwork/corelog"
	rpc "github.com/sourcenetwork/go-libp2p-pubsub-rpc"
	"github.com/sourcenetwork/immutable"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpcpeer "google.golang.org/grpc/peer"
	"google.golang.org/protobuf/proto"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/errors"
	pb "github.com/sourcenetwork/defradb/net/pb"
)

// Server is the request/response instance for all P2P RPC communication.
// Implements gRPC server. See net/pb/net.proto for corresponding service definitions.
//
// Specifically, server handles the push/get request/response aspects of the RPC service
// but not the API calls.
type server struct {
	peer *Peer
	opts []grpc.DialOption
	db   client.DB

	topics map[string]pubsubTopic
	mu     sync.Mutex

	conns map[libpeer.ID]*grpc.ClientConn

	pubSubEmitter  event.Emitter
	pushLogEmitter event.Emitter

	// docQueue is used to track which documents are currently being processed.
	// This is used to prevent multiple concurrent processing of the same document and
	// limit unecessary transaction conflicts.
	docQueue *docQueue

	pb.UnimplementedServiceServer
}

// pubsubTopic is a wrapper of rpc.Topic to be able to track if the topic has
// been subscribed to.
type pubsubTopic struct {
	*rpc.Topic
	subscribed bool
}

// newServer creates a new network server that handle/directs RPC requests to the
// underlying DB instance.
func newServer(p *Peer, db client.DB, opts ...grpc.DialOption) (*server, error) {
	s := &server{
		peer:   p,
		conns:  make(map[libpeer.ID]*grpc.ClientConn),
		topics: make(map[string]pubsubTopic),
		db:     db,
		docQueue: &docQueue{
			docs: make(map[string]chan struct{}),
		},
	}

	cred := insecure.NewCredentials()
	defaultOpts := []grpc.DialOption{
		s.getLibp2pDialer(),
		grpc.WithTransportCredentials(cred),
	}

	s.opts = append(defaultOpts, opts...)
	if s.peer.ps != nil {
		colMap, err := p.loadP2PCollections(p.ctx)
		if err != nil {
			return nil, err
		}

		// Get all DocIDs across all collections in the DB
		cols, err := s.db.GetCollections(s.peer.ctx, client.CollectionFetchOptions{})
		if err != nil {
			return nil, err
		}

		i := 0
		for _, col := range cols {
			// If we subscribed to the collection, we skip subscribing to the collection's docIDs.
			if _, ok := colMap[col.SchemaRoot()]; ok {
				continue
			}
			docIDChan, err := col.GetAllDocIDs(p.ctx)
			if err != nil {
				return nil, err
			}

			for docID := range docIDChan {
				if err := s.addPubSubTopic(docID.ID.String(), true); err != nil {
					return nil, err
				}
				i++
			}
		}
	}

	var err error
	s.pubSubEmitter, err = s.peer.host.EventBus().Emitter(new(EvtPubSub))
	if err != nil {
		log.InfoContext(s.peer.ctx, "could not create event emitter", corelog.String("Error", err.Error()))
	}
	s.pushLogEmitter, err = s.peer.host.EventBus().Emitter(new(EvtReceivedPushLog))
	if err != nil {
		log.InfoContext(s.peer.ctx, "could not create event emitter", corelog.String("Error", err.Error()))
	}

	return s, nil
}

// GetDocGraph receives a get graph request
func (s *server) GetDocGraph(
	ctx context.Context,
	req *pb.GetDocGraphRequest,
) (*pb.GetDocGraphReply, error) {
	return nil, nil
}

// PushDocGraph receives a push graph request
func (s *server) PushDocGraph(
	ctx context.Context,
	req *pb.PushDocGraphRequest,
) (*pb.PushDocGraphReply, error) {
	return nil, nil
}

// GetLog receives a get log request
func (s *server) GetLog(ctx context.Context, req *pb.GetLogRequest) (*pb.GetLogReply, error) {
	return nil, nil
}

type docQueue struct {
	docs map[string]chan struct{}
	mu   sync.Mutex
}

// add adds a docID to the queue. If the docID is already in the queue, it will
// wait for the docID to be removed from the queue. For every add call, done must
// be called to remove the docID from the queue. Otherwise, subsequent add calls will
// block forever.
func (dq *docQueue) add(docID string) {
	dq.mu.Lock()
	done, ok := dq.docs[docID]
	if !ok {
		dq.docs[docID] = make(chan struct{})
	}
	dq.mu.Unlock()
	if ok {
		<-done
		dq.add(docID)
	}
}

func (dq *docQueue) done(docID string) {
	dq.mu.Lock()
	defer dq.mu.Unlock()
	done, ok := dq.docs[docID]
	if ok {
		delete(dq.docs, docID)
		close(done)
	}
}

// PushLog receives a push log request
func (s *server) PushLog(ctx context.Context, req *pb.PushLogRequest) (*pb.PushLogReply, error) {
	pid, err := peerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	cid, err := cid.Cast(req.Body.Cid)
	if err != nil {
		return nil, err
	}
	docID, err := client.NewDocIDFromString(string(req.Body.DocID))
	if err != nil {
		return nil, err
	}

	s.docQueue.add(docID.String())
	defer func() {
		s.docQueue.done(docID.String())
		if s.pushLogEmitter != nil {
			byPeer, err := libpeer.Decode(req.Body.Creator)
			if err != nil {
				log.InfoContext(ctx, "could not decode the PeerID of the log creator", corelog.String("Error", err.Error()))
			}
			err = s.pushLogEmitter.Emit(EvtReceivedPushLog{
				FromPeer: pid,
				ByPeer:   byPeer,
			})
			if err != nil {
				// logging instead of returning an error because the event bus should
				// not break the PushLog execution.
				log.InfoContext(ctx, "could not emit push log event", corelog.String("Error", err.Error()))
			}
		}
	}()

	// make sure were not processing twice
	if canVisit := s.peer.queuedChildren.Visit(cid); !canVisit {
		return &pb.PushLogReply{}, nil
	}
	defer s.peer.queuedChildren.Remove(cid)

	// check if we already have this block
	exists, err := s.db.Blockstore().Has(ctx, cid)
	if err != nil {
		return nil, errors.Wrap(fmt.Sprintf("failed to check for existing block %s", cid), err)
	}
	if exists {
		return &pb.PushLogReply{}, nil
	}

	dsKey := core.DataStoreKeyFromDocID(docID)

	var txnErr error
	for retry := 0; retry < s.peer.db.MaxTxnRetries(); retry++ {
		// To prevent a potential deadlock on DAG sync if an error occures mid process, we handle
		// each process on a single transaction.
		txn, err := s.db.NewConcurrentTxn(ctx, false)
		if err != nil {
			return nil, err
		}
		defer txn.Discard(ctx)
		store := s.db.WithTxn(txn)

		// Currently a schema is the best way we have to link a push log request to a collection,
		// this will change with https://github.com/sourcenetwork/defradb/issues/1085
		col, err := s.getActiveCollection(ctx, store, string(req.Body.SchemaRoot))
		if err != nil {
			return nil, err
		}

		// Create a new DAG service with the current transaction
		var getter format.NodeGetter = s.peer.newDAGSyncerTxn(txn)
		if sessionMaker, ok := getter.(SessionDAGSyncer); ok {
			getter = sessionMaker.Session(ctx)
		}

		// handleComposite
		nd, err := decodeBlockBuffer(req.Body.Log.Block, cid)
		if err != nil {
			return nil, errors.Wrap("failed to decode block to ipld.Node", err)
		}

		var session sync.WaitGroup
		bp := newBlockProcessor(s.peer, txn, col, dsKey, getter)
		err = bp.processRemoteBlock(ctx, &session, nd, true)
		if err != nil {
			log.ErrorContextE(
				ctx,
				"Failed to process remote block",
				err,
				corelog.String("DocID", dsKey.DocID),
				corelog.Any("CID", cid),
			)
		}
		session.Wait()
		bp.mergeBlocks(ctx)

		err = s.syncIndexedDocs(ctx, col.WithTxn(txn), docID)
		if err != nil {
			return nil, err
		}

		// dagWorkers specific to the DocID will have been spawned within handleChildBlocks.
		// Once we are done with the dag syncing process, we can get rid of those workers.
		if s.peer.closeJob != nil {
			s.peer.closeJob <- dsKey.DocID
		}

		if txnErr = txn.Commit(ctx); txnErr != nil {
			if errors.Is(txnErr, badger.ErrTxnConflict) {
				continue
			}
			return &pb.PushLogReply{}, txnErr
		}

		// Once processed, subscribe to the DocID topic on the pubsub network unless we already
		// suscribe to the collection.
		if !s.hasPubSubTopic(col.SchemaRoot()) {
			err = s.addPubSubTopic(dsKey.DocID, true)
			if err != nil {
				return nil, err
			}
		}
		return &pb.PushLogReply{}, nil
	}

	return &pb.PushLogReply{}, client.NewErrMaxTxnRetries(txnErr)
}

func (*server) getActiveCollection(
	ctx context.Context,
	store client.Store,
	schemaRoot string,
) (client.Collection, error) {
	cols, err := store.GetCollections(
		ctx,
		client.CollectionFetchOptions{
			SchemaRoot: immutable.Some(schemaRoot),
		},
	)
	if err != nil {
		return nil, errors.Wrap(fmt.Sprintf("Failed to get collection from schemaRoot %s", schemaRoot), err)
	}
	if len(cols) == 0 {
		return nil, client.NewErrCollectionNotFoundForSchema(schemaRoot)
	}
	var col client.Collection
	for _, c := range cols {
		if col != nil && col.Name().HasValue() && !c.Name().HasValue() {
			continue
		}
		col = c
	}
	return col, nil
}

func (s *server) syncIndexedDocs(
	ctx context.Context,
	col client.Collection,
	docID client.DocID,
) error {
	preTxnCol, err := s.db.GetCollectionByName(ctx, col.Name().Value())
	if err != nil {
		return err
	}

	//TODO-ACP: https://github.com/sourcenetwork/defradb/issues/2365
	// Resolve while handling acp <> secondary indexes.
	oldDoc, err := preTxnCol.Get(ctx, acpIdentity.NoIdentity, docID, false)
	isNewDoc := errors.Is(err, client.ErrDocumentNotFound)
	if !isNewDoc && err != nil {
		return err
	}

	//TODO-ACP: https://github.com/sourcenetwork/defradb/issues/2365
	// Resolve while handling acp <> secondary indexes.
	doc, err := col.Get(ctx, acpIdentity.NoIdentity, docID, false)
	isDeletedDoc := errors.Is(err, client.ErrDocumentNotFound)
	if !isDeletedDoc && err != nil {
		return err
	}

	if isDeletedDoc {
		return preTxnCol.DeleteDocIndex(ctx, oldDoc)
	} else if isNewDoc {
		return col.CreateDocIndex(ctx, doc)
	} else {
		return col.UpdateDocIndex(ctx, oldDoc, doc)
	}
}

// GetHeadLog receives a get head log request
func (s *server) GetHeadLog(
	ctx context.Context,
	req *pb.GetHeadLogRequest,
) (*pb.GetHeadLogReply, error) {
	return nil, nil
}

// addPubSubTopic subscribes to a topic on the pubsub network
func (s *server) addPubSubTopic(topic string, subscribe bool) error {
	if s.peer.ps == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.topics[topic]; ok {
		// When the topic was previously set to publish only and we now want to subscribe,
		// we need to close the existing topic and create a new one.
		if !t.subscribed && subscribe {
			if err := t.Close(); err != nil {
				return err
			}
		} else {
			return nil
		}
	}

	t, err := rpc.NewTopic(s.peer.ctx, s.peer.ps, s.peer.host.ID(), topic, subscribe)
	if err != nil {
		return err
	}

	t.SetEventHandler(s.pubSubEventHandler)
	t.SetMessageHandler(s.pubSubMessageHandler)
	s.topics[topic] = pubsubTopic{
		Topic:      t,
		subscribed: subscribe,
	}
	return nil
}

// hasPubSubTopic checks if we are subscribed to a topic.
func (s *server) hasPubSubTopic(topic string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.topics[topic]
	return ok
}

// removePubSubTopic unsubscribes to a topic
func (s *server) removePubSubTopic(topic string) error {
	if s.peer.ps == nil {
		return nil
	}

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
func (s *server) publishLog(ctx context.Context, topic string, req *pb.PushLogRequest) error {
	if s.peer.ps == nil { // skip if we aren't running with a pubsub net
		return nil
	}
	s.mu.Lock()
	t, ok := s.topics[topic]
	s.mu.Unlock()
	if !ok {
		err := s.addPubSubTopic(topic, false)
		if err != nil {
			return errors.Wrap(fmt.Sprintf("failed to created single use topic %s", topic), err)
		}
		return s.publishLog(ctx, topic, req)
	}

	data, err := req.MarshalVT()
	if err != nil {
		return errors.Wrap("failed marshling pubsub message", err)
	}

	if _, err := t.Publish(ctx, data, rpc.WithIgnoreResponse(true)); err != nil {
		return errors.Wrap(fmt.Sprintf("failed publishing to thread %s", topic), err)
	}
	return nil
}

// pubSubMessageHandler handles incoming PushLog messages from the pubsub network.
func (s *server) pubSubMessageHandler(from libpeer.ID, topic string, msg []byte) ([]byte, error) {
	req := new(pb.PushLogRequest)
	if err := proto.Unmarshal(msg, req); err != nil {
		log.ErrorContextE(s.peer.ctx, "Failed to unmarshal pubsub message %s", err)
		return nil, err
	}

	ctx := grpcpeer.NewContext(s.peer.ctx, &grpcpeer.Peer{
		Addr: addr{from},
	})
	if _, err := s.PushLog(ctx, req); err != nil {
		return nil, errors.Wrap(fmt.Sprintf("Failed pushing log for doc %s", topic), err)
	}
	return nil, nil
}

// pubSubEventHandler logs events from the subscribed DocID topics.
func (s *server) pubSubEventHandler(from libpeer.ID, topic string, msg []byte) {
	log.InfoContext(
		s.peer.ctx,
		"Received new pubsub event",
		corelog.Any("SenderId", from),
		corelog.String("Topic", topic),
		corelog.String("Message", string(msg)),
	)

	if s.pubSubEmitter != nil {
		err := s.pubSubEmitter.Emit(EvtPubSub{
			Peer: from,
		})
		if err != nil {
			log.InfoContext(s.peer.ctx, "could not emit pubsub event", corelog.Any("Error", err.Error()))
		}
	}
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

// KEEPING AS REFERENCE
//
// logFromProto returns a thread log from a proto log.
// func logFromProto(l *pb.Log) thread.LogInfo {
// 	return thread.LogInfo{
// 		ID:     l.ID.ID,
// 		PubKey: l.PubKey.PubKey,
// 		Addrs:  addrsFromProto(l.Addrs),
// 		Head: thread.Head{
// 			ID:      l.Head.Cid,
// 			Counter: l.Counter,
// 		},
// 	}
// }
