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

	"github.com/gogo/protobuf/proto"
	format "github.com/ipfs/go-ipld-format"
	"github.com/libp2p/go-libp2p-core/event"
	libpeer "github.com/libp2p/go-libp2p-core/peer"
	rpc "github.com/textileio/go-libp2p-pubsub-rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpcpeer "google.golang.org/grpc/peer"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/logging"
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

	topics map[string]*rpc.Topic
	mu     sync.Mutex

	conns map[libpeer.ID]*grpc.ClientConn

	pubSubEmitter  event.Emitter
	pushLogEmitter event.Emitter
}

// newServer creates a new network server that handle/directs RPC requests to the
// underlying DB instance.
func newServer(p *Peer, db client.DB, opts ...grpc.DialOption) (*server, error) {
	s := &server{
		peer:   p,
		conns:  make(map[libpeer.ID]*grpc.ClientConn),
		topics: make(map[string]*rpc.Topic),
		db:     db,
	}

	cred := insecure.NewCredentials()
	defaultOpts := []grpc.DialOption{
		s.getLibp2pDialer(),
		grpc.WithTransportCredentials(cred),
	}

	s.opts = append(defaultOpts, opts...)
	if s.peer.ps != nil {
		// Get all DocKeys across all collections in the DB
		log.Debug(p.ctx, "Getting all existing DocKey...")
		keyResults, err := s.listAllDocKeys()
		if err != nil {
			return nil, fmt.Errorf("Failed to get DocKeys for pubsub topic registration: %w", err)
		}

		i := 0
		if keyResults != nil {
			for key := range keyResults {
				if key.Err != nil {
					log.ErrorE(p.ctx, "Failed to get a key to register pubsub topic", key.Err)
					continue
				}
				log.Debug(
					p.ctx,
					"Registering existing DocKey pubsub topic",
					logging.NewKV("DocKey", key.Key.String()),
				)
				if err := s.addPubSubTopic(key.Key.String()); err != nil {
					return nil, err
				}
				i++
			}
		}
		log.Debug(p.ctx, "Finished registering all DocKey pubsub topics", logging.NewKV("Count", i))
	}

	var err error
	s.pubSubEmitter, err = s.peer.host.EventBus().Emitter(new(EvtPubSub))
	if err != nil {
		log.Info(s.peer.ctx, "could not create event emitter", logging.NewKV("Error", err))
	}
	s.pushLogEmitter, err = s.peer.host.EventBus().Emitter(new(EvtReceivedPushLog))
	if err != nil {
		log.Info(s.peer.ctx, "could not create event emitter", logging.NewKV("Error", err))
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

// PushLog receives a push log request
func (s *server) PushLog(ctx context.Context, req *pb.PushLogRequest) (*pb.PushLogReply, error) {
	pid, err := peerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	log.Debug(ctx, "Received a PushLog request", logging.NewKV("PID", pid))

	// parse request object
	cid := req.Body.Cid.Cid

	// make sure were not processing twice
	if canVisit := s.peer.queuedChildren.Visit(cid); !canVisit {
		return &pb.PushLogReply{}, nil
	}

	schemaID := string(req.Body.SchemaID)
	docKey := core.DataStoreKeyFromDocKey(req.Body.DocKey.DocKey)
	col, err := s.db.GetCollectionBySchemaID(ctx, schemaID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get collection from schemaID %s: %w", schemaID, err)
	}

	var getter format.NodeGetter = s.peer.ds
	if sessionMaker, ok := getter.(SessionDAGSyncer); ok {
		log.Debug(ctx, "Upgrading DAGSyncer with a session")
		getter = sessionMaker.Session(ctx)
	}

	// handleComposite
	nd, err := decodeBlockBuffer(req.Body.Log.Block, cid)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode block to ipld.Node: %w", err)
	}
	cids, err := s.peer.processLog(ctx, col, docKey, cid, "", nd, getter)
	if err != nil {
		log.ErrorE(
			ctx,
			"Failed to process PushLog node",
			err,
			logging.NewKV("DocKey", docKey),
			logging.NewKV("CID", cid),
		)
	}

	// handleChildren
	if len(cids) > 0 { // we have child nodes to get
		log.Debug(
			ctx,
			"Handling children for log",
			logging.NewKV("NChildren", len(cids)),
			logging.NewKV("CID", cid),
		)
		var session sync.WaitGroup
		s.peer.handleChildBlocks(&session, col, docKey, "", nd, cids, getter)
		session.Wait()
	} else {
		log.Debug(ctx, "No more children to process for log", logging.NewKV("CID", cid))
	}

	if s.pushLogEmitter != nil {
		err = s.pushLogEmitter.Emit(EvtReceivedPushLog{
			Peer: pid,
		})
		if err != nil {
			log.Debug(ctx, "could not emit push log event", logging.NewKV("Error", err))
		}
	}

	return &pb.PushLogReply{}, nil
}

// GetHeadLog receives a get head log request
func (s *server) GetHeadLog(
	ctx context.Context,
	req *pb.GetHeadLogRequest,
) (*pb.GetHeadLogReply, error) {
	return nil, nil
}

// addPubSubTopic subscribes to a DocKey topic
func (s *server) addPubSubTopic(dockey string) error {
	if s.peer.ps == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.topics[dockey]; ok {
		return nil
	}

	t, err := rpc.NewTopic(s.peer.ctx, s.peer.ps, s.peer.host.ID(), dockey, true)
	if err != nil {
		return err
	}

	t.SetEventHandler(s.pubSubEventHandler)
	t.SetMessageHandler(s.pubSubMessageHandler)
	s.topics[dockey] = t
	return nil
}

// removePubSubTopic unsubscribes to a DocKey topic
//nolint:unused
func (s *server) removePubSubTopic(dockey string) error {
	if s.peer.ps == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.topics[dockey]; ok {
		delete(s.topics, dockey)
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
func (s *server) publishLog(ctx context.Context, dockey string, req *pb.PushLogRequest) error {
	if s.peer.ps == nil { // skip if we aren't running with a pubsub net
		return nil
	}
	s.mu.Lock()
	t, ok := s.topics[dockey]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("No pubsub topic found for doc %s", dockey)
	}

	data, err := req.Marshal()
	if err != nil {
		return fmt.Errorf("failed marshling pubsub message: %w", err)
	}

	if _, err := t.Publish(ctx, data, rpc.WithIgnoreResponse(true)); err != nil {
		return fmt.Errorf("failed publishing to thread %s: %w", dockey, err)
	}
	log.Debug(
		ctx,
		"Published log",
		logging.NewKV("CID", req.Body.Cid.Cid),
		logging.NewKV("DocKey", dockey),
	)
	return nil
}

// pubSubMessageHandler handles incoming PushLog messages from the pubsub network.
func (s *server) pubSubMessageHandler(from libpeer.ID, topic string, msg []byte) ([]byte, error) {
	log.Debug(
		s.peer.ctx,
		"Handling new pubsub message",
		logging.NewKV("SenderId", from),
		logging.NewKV("Topic", topic),
	)
	req := new(pb.PushLogRequest)
	if err := proto.Unmarshal(msg, req); err != nil {
		log.ErrorE(s.peer.ctx, "Failed to unmarshal pubsub message %s", err)
		return nil, err
	}

	ctx := grpcpeer.NewContext(s.peer.ctx, &grpcpeer.Peer{
		Addr: addr{from},
	})
	if _, err := s.PushLog(ctx, req); err != nil {
		log.ErrorE(ctx, "Failed pushing log for doc", err, logging.NewKV("Topic", topic))
		return nil, fmt.Errorf("Failed pushing log for doc %s: %w", topic, err)
	}
	return nil, nil
}

// pubSubEventHandler logs events from the subscribed dockey topics.
func (s *server) pubSubEventHandler(from libpeer.ID, topic string, msg []byte) {
	log.Info(
		s.peer.ctx,
		"Received new pubsub event",
		logging.NewKV("SenderId", from),
		logging.NewKV("Topic", topic),
		logging.NewKV("Message", string(msg)),
	)

	if s.pubSubEmitter != nil {
		err := s.pubSubEmitter.Emit(EvtPubSub{
			Peer: from,
		})
		if err != nil {
			log.Info(s.peer.ctx, "could not emit pubsub event", logging.NewKV("Error", err))
		}
	}
}

func (s *server) listAllDocKeys() (<-chan client.DocKeysResult, error) {
	// get all collections
	cols, err := s.db.GetAllCollections(s.peer.ctx)
	if err != nil {
		return nil, err
	}

	if len(cols) == 0 {
		return nil, nil
	}

	keyCh := make(chan client.DocKeysResult)

	var wg sync.WaitGroup
	wg.Add(1) // add an init blocker on close routine
	go func() {
		wg.Wait()
		close(keyCh)
	}()

	for _, col := range cols {
		resCh, err := col.GetAllDocKeys(s.peer.ctx)
		if err != nil {
			return nil, err
		}

		// run a goroutine for each channel we get from the GetAllDocKeys func for each
		// collection. Pipe the results from res to keys, and handle potentially
		// closed channel edge cases
		wg.Add(1)
		go func(colName string) {
			for res := range resCh {
				keyCh <- res
			}
			wg.Done()
		}(col.Name())
	}
	wg.Done() // cleanup the init blocker on close routine

	return keyCh, nil
}

// addr implements net.Addr and holds a libp2p peer ID.
type addr struct{ id libpeer.ID }

// Network returns the name of the network that this address belongs to (libp2p).
func (a addr) Network() string { return "libp2p" }

// String returns the peer ID of this address in string form (B58-encoded).
func (a addr) String() string { return a.id.Pretty() }

// peerIDFromContext returns peer ID from the GRPC context
func peerIDFromContext(ctx context.Context) (libpeer.ID, error) {
	ctxPeer, ok := grpcpeer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("unable to identify stream peer")
	}
	pid, err := libpeer.Decode(ctxPeer.Addr.String())
	if err != nil {
		return "", fmt.Errorf("parsing stream peer id: %w", err)
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
