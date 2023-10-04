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

	"github.com/ipfs/boxo/bitswap"
	"github.com/ipfs/boxo/bitswap/network"
	"github.com/ipfs/boxo/blockservice"
	exchange "github.com/ipfs/boxo/exchange"
	dag "github.com/ipfs/boxo/ipld/merkledag"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	ipld "github.com/ipfs/go-ipld-format"
	gostream "github.com/libp2p/go-libp2p-gostream"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	peerstore "github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/routing"
	"google.golang.org/grpc"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	corenet "github.com/sourcenetwork/defradb/core/net"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/merkle/clock"
	pb "github.com/sourcenetwork/defradb/net/pb"
)

var (
	numWorkers = 5
)

// Peer is a DefraDB Peer node which exposes all the LibP2P host/peer functionality
// to the underlying DefraDB instance.
type Peer struct {
	//config??

	db            client.DB
	updateChannel chan events.Update

	host host.Host
	dht  routing.Routing
	ps   *pubsub.PubSub

	server *server
	p2pRPC *grpc.Server // rpc server over the P2P network

	// Used to close the dagWorker pool for a given document.
	// The string represents a dockey.
	closeJob chan string
	sendJobs chan *dagJob

	// outstanding log request currently being processed
	queuedChildren *cidSafeSet

	// replicators is a map from collectionName => peerId
	replicators map[string]map[peer.ID]struct{}
	mu          sync.Mutex

	// peer DAG service
	ipld.DAGService
	exch  exchange.Interface
	bserv blockservice.BlockService

	ctx    context.Context
	cancel context.CancelFunc
}

// NewPeer creates a new instance of the DefraDB server as a peer-to-peer node.
func NewPeer(
	ctx context.Context,
	db client.DB,
	h host.Host,
	dht routing.Routing,
	ps *pubsub.PubSub,
	serverOptions []grpc.ServerOption,
	dialOptions []grpc.DialOption,
) (*Peer, error) {
	if db == nil {
		return nil, ErrNilDB
	}

	ctx, cancel := context.WithCancel(ctx)
	p := &Peer{
		host:           h,
		dht:            dht,
		ps:             ps,
		db:             db,
		p2pRPC:         grpc.NewServer(serverOptions...),
		ctx:            ctx,
		cancel:         cancel,
		closeJob:       make(chan string),
		sendJobs:       make(chan *dagJob),
		replicators:    make(map[string]map[peer.ID]struct{}),
		queuedChildren: newCidSafeSet(),
	}
	var err error
	p.server, err = newServer(p, db, dialOptions...)
	if err != nil {
		return nil, err
	}

	err = p.loadReplicators(p.ctx)
	if err != nil {
		return nil, err
	}

	p.setupBlockService()
	p.setupDAGService()

	return p, nil
}

// Start all the internal workers/goroutines/loops that manage the P2P state.
func (p *Peer) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// reconnect to known peers
	var wg sync.WaitGroup
	for _, id := range p.host.Peerstore().PeersWithAddrs() {
		if id == p.host.ID() {
			continue
		}
		wg.Add(1)
		go func(id peer.ID) {
			defer wg.Done()
			addr := p.host.Peerstore().PeerInfo(id)
			err := p.host.Connect(p.ctx, addr)
			if err != nil {
				log.Info(
					p.ctx,
					"Failure while reconnecting to a known peer",
					logging.NewKV("peer", id),
					logging.NewKV("error", err),
				)
			}
		}(id)
	}
	wg.Wait()

	p2plistener, err := gostream.Listen(p.host, corenet.Protocol)
	if err != nil {
		return err
	}

	if p.ps != nil {
		if !p.db.Events().Updates.HasValue() {
			return ErrNilUpdateChannel
		}

		updateChannel, err := p.db.Events().Updates.Value().Subscribe()
		if err != nil {
			return err
		}
		p.updateChannel = updateChannel

		log.Info(p.ctx, "Starting internal broadcaster for pubsub network")
		go p.handleBroadcastLoop()
	}

	// register the P2P gRPC server
	go func() {
		pb.RegisterServiceServer(p.p2pRPC, p.server)
		if err := p.p2pRPC.Serve(p2plistener); err != nil &&
			!errors.Is(err, grpc.ErrServerStopped) {
			log.FatalE(p.ctx, "Fatal P2P RPC server error", err)
		}
	}()

	// start sendJobWorker
	go p.sendJobWorker()

	return nil
}

// Close the peer node and all its internal workers/goroutines/loops.
func (p *Peer) Close() error {
	// close topics
	if err := p.server.removeAllPubsubTopics(); err != nil {
		log.ErrorE(p.ctx, "Error closing pubsub topics", err)
	}

	// stop gRPC server
	for _, c := range p.server.conns {
		if err := c.Close(); err != nil {
			log.ErrorE(p.ctx, "Failed closing server RPC connections", err)
		}
	}
	stopGRPCServer(p.ctx, p.p2pRPC)
	// stopGRPCServer(p.tcpRPC)

	// close event emitters
	if p.server.pubSubEmitter != nil {
		if err := p.server.pubSubEmitter.Close(); err != nil {
			log.Info(p.ctx, "Could not close pubsub event emitter", logging.NewKV("Error", err.Error()))
		}
	}
	if p.server.pushLogEmitter != nil {
		if err := p.server.pushLogEmitter.Close(); err != nil {
			log.Info(p.ctx, "Could not close push log event emitter", logging.NewKV("Error", err.Error()))
		}
	}

	if p.db.Events().Updates.HasValue() {
		p.db.Events().Updates.Value().Unsubscribe(p.updateChannel)
	}

	if err := p.bserv.Close(); err != nil {
		log.ErrorE(p.ctx, "Error closing block service", err)
	}

	if err := p.host.Close(); err != nil {
		log.ErrorE(p.ctx, "Error closing host", err)
	}

	p.cancel()
	return nil
}

// handleBroadcast loop manages the transition of messages
// from the internal broadcaster to the external pubsub network
func (p *Peer) handleBroadcastLoop() {
	log.Debug(p.ctx, "Waiting for messages on internal broadcaster")
	for {
		log.Debug(p.ctx, "Handling internal broadcast bus message")
		update, isOpen := <-p.updateChannel
		if !isOpen {
			return
		}

		// check log priority, 1 is new doc log
		// 2 is update log
		var err error
		if update.Priority == 1 {
			err = p.handleDocCreateLog(update)
		} else if update.Priority > 1 {
			err = p.handleDocUpdateLog(update)
		} else {
			log.Info(p.ctx, "Skipping log with invalid priority of 0", logging.NewKV("CID", update.Cid))
		}

		if err != nil {
			log.ErrorE(p.ctx, "Error while handling broadcast log", err)
		}
	}
}

// RegisterNewDocument registers a new document with the peer node.
func (p *Peer) RegisterNewDocument(
	ctx context.Context,
	dockey client.DocKey,
	c cid.Cid,
	nd ipld.Node,
	schemaID string,
) error {
	log.Debug(
		p.ctx,
		"Registering a new document for our peer node",
		logging.NewKV("DocKey", dockey.String()),
	)

	// register topic
	if err := p.server.addPubSubTopic(dockey.String(), !p.server.hasPubSubTopic(schemaID)); err != nil {
		log.ErrorE(
			p.ctx,
			"Failed to create new pubsub topic",
			err,
			logging.NewKV("DocKey", dockey.String()),
		)
		return err
	}

	// publish log
	body := &pb.PushLogRequest_Body{
		DocKey:   []byte(dockey.String()),
		Cid:      c.Bytes(),
		SchemaID: []byte(schemaID),
		Creator:  p.host.ID().String(),
		Log: &pb.Document_Log{
			Block: nd.RawData(),
		},
	}
	req := &pb.PushLogRequest{
		Body: body,
	}

	return p.server.publishLog(p.ctx, schemaID, req)
}

// SetReplicator adds a target peer node as a replication destination for documents in our DB.
func (p *Peer) SetReplicator(
	ctx context.Context,
	rep client.Replicator,
) error {
	txn, err := p.db.NewTxn(ctx, true)
	if err != nil {
		return err
	}
	store := p.db.WithTxn(txn)

	err = p.setReplicator(ctx, store, rep.Info, rep.Schemas...)
	if err != nil {
		txn.Discard(ctx)
		return err
	}

	return txn.Commit(ctx)
}

// setReplicator adds a target peer node as a replication destination for documents in our DB.
func (p *Peer) setReplicator(
	ctx context.Context,
	store client.Store,
	info peer.AddrInfo,
	collectionNames ...string,
) error {
	// verify collections
	collections := []client.Collection{}
	schemas := []string{}
	if len(collectionNames) == 0 {
		var err error
		collections, err = store.GetAllCollections(ctx)
		if err != nil {
			return errors.Wrap("failed to get all collections for replicator", err)
		}
		for _, col := range collections {
			schemas = append(schemas, col.SchemaID())
		}
	} else {
		for _, cName := range collectionNames {
			col, err := store.GetCollectionByName(ctx, cName)
			if err != nil {
				return errors.Wrap("failed to get collection for replicator", err)
			}
			collections = append(collections, col)
			schemas = append(schemas, col.SchemaID())
		}
	}

	// make sure it's not ourselves
	if info.ID == p.host.ID() {
		return errors.New("can't target ourselves as a replicator")
	}
	if err := info.ID.Validate(); err != nil {
		return err
	}

	// Add the destination's peer multiaddress in the peerstore.
	// This will be used during connection and stream creation by libp2p.
	p.host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	// make sure we're not duplicating things
	p.mu.Lock()
	for _, col := range collections {
		if reps, exists := p.replicators[col.SchemaID()]; exists {
			if _, exists := reps[info.ID]; exists {
				p.mu.Unlock()
				return errors.New(fmt.Sprintf(
					"Replicator already exists for %s with PeerID %s",
					col.Name(),
					info.ID,
				))
			}
		} else {
			p.replicators[col.SchemaID()] = make(map[peer.ID]struct{})
		}
		// add to replicators list for the collection
		p.replicators[col.SchemaID()][info.ID] = struct{}{}
	}
	p.mu.Unlock()

	// Persist peer in datastore
	err := p.db.SetReplicator(ctx, client.Replicator{
		Info:    info,
		Schemas: schemas,
	})
	if err != nil {
		return errors.Wrap("failed to persist replicator", err)
	}

	for _, col := range collections {
		// create read only txn and assign to col
		txn, err := p.db.NewTxn(ctx, true)
		if err != nil {
			return errors.Wrap("failed to get txn", err)
		}
		col = col.WithTxn(txn)

		// get dockeys (all)
		keysCh, err := col.GetAllDocKeys(ctx)
		if err != nil {
			txn.Discard(ctx)
			return errors.Wrap(
				fmt.Sprintf(
					"Failed to get dockey for replicator %s on %s",
					info.ID,
					col.Name(),
				),
				err,
			)
		}

		p.pushToReplicator(ctx, txn, col, keysCh, info.ID)
	}
	return nil
}

func (p *Peer) pushToReplicator(
	ctx context.Context,
	txn datastore.Txn,
	collection client.Collection,
	keysCh <-chan client.DocKeysResult,
	pid peer.ID,
) {
	for key := range keysCh {
		if key.Err != nil {
			log.ErrorE(ctx, "Key channel error", key.Err)
			continue
		}
		dockey := core.DataStoreKeyFromDocKey(key.Key)
		headset := clock.NewHeadSet(
			txn.Headstore(),
			dockey.WithFieldId(core.COMPOSITE_NAMESPACE).ToHeadStoreKey(),
		)
		cids, priority, err := headset.List(ctx)
		if err != nil {
			log.ErrorE(
				ctx,
				"Failed to get heads",
				err,
				logging.NewKV("DocKey", key.Key.String()),
				logging.NewKV("PeerID", pid),
				logging.NewKV("Collection", collection.Name()))
			continue
		}
		// loop over heads, get block, make the required logs, and send
		for _, c := range cids {
			blk, err := txn.DAGstore().Get(ctx, c)
			if err != nil {
				log.ErrorE(ctx, "Failed to get block", err,
					logging.NewKV("CID", c),
					logging.NewKV("PeerID", pid),
					logging.NewKV("Collection", collection.Name()))
				continue
			}

			// @todo: remove encode/decode loop for core.Log data
			nd, err := dag.DecodeProtobuf(blk.RawData())
			if err != nil {
				log.ErrorE(ctx, "Failed to decode protobuf", err, logging.NewKV("CID", c))
				continue
			}

			evt := events.Update{
				DocKey:   key.Key.String(),
				Cid:      c,
				SchemaID: collection.SchemaID(),
				Block:    nd,
				Priority: priority,
			}
			if err := p.server.pushLog(ctx, evt, pid); err != nil {
				log.ErrorE(
					ctx,
					"Failed to replicate log",
					err,
					logging.NewKV("CID", c),
					logging.NewKV("PeerID", pid),
				)
			}
		}
	}
}

// DeleteReplicator removes a peer node from the replicators.
func (p *Peer) DeleteReplicator(
	ctx context.Context,
	rep client.Replicator,
) error {
	log.Debug(ctx, "Received DeleteReplicator request")

	txn, err := p.db.NewTxn(ctx, true)
	if err != nil {
		return err
	}
	store := p.db.WithTxn(txn)

	err = p.deleteReplicator(ctx, store, rep.Info, rep.Schemas...)
	if err != nil {
		txn.Discard(ctx)
		return err
	}

	return txn.Commit(ctx)
}

func (p *Peer) deleteReplicator(
	ctx context.Context,
	store client.Store,
	info peer.AddrInfo,
	collectionNames ...string,
) error {
	// make sure it's not ourselves
	if info.ID == p.host.ID() {
		return ErrSelfTargetForReplicator
	}
	if err := info.ID.Validate(); err != nil {
		return err
	}

	// verify collections
	schemas := []string{}
	schemaMap := make(map[string]struct{})
	if len(collectionNames) == 0 {
		var err error
		collections, err := store.GetAllCollections(ctx)
		if err != nil {
			return errors.Wrap("failed to get all collections for replicator", err)
		}
		for _, col := range collections {
			schemas = append(schemas, col.SchemaID())
			schemaMap[col.SchemaID()] = struct{}{}
		}
	} else {
		for _, cName := range collectionNames {
			col, err := store.GetCollectionByName(ctx, cName)
			if err != nil {
				return errors.Wrap("failed to get collection for replicator", err)
			}
			schemas = append(schemas, col.SchemaID())
			schemaMap[col.SchemaID()] = struct{}{}
		}
	}

	// make sure we're not duplicating things
	p.mu.Lock()
	defer p.mu.Unlock()

	totalSchemas := 0 // Lets keep track of how many schemas are left for the replicator.
	for schema, rep := range p.replicators {
		if _, exists := rep[info.ID]; exists {
			if _, toDelete := schemaMap[schema]; toDelete {
				delete(p.replicators[schema], info.ID)
			} else {
				totalSchemas++
			}
		}
	}

	if totalSchemas == 0 {
		// Remove the destination's peer multiaddress in the peerstore.
		p.host.Peerstore().ClearAddrs(info.ID)
	}

	// Delete peer in datastore
	return p.db.DeleteReplicator(ctx, client.Replicator{
		Info:    peer.AddrInfo{ID: info.ID},
		Schemas: schemas,
	})
}

// GetAllReplicators returns all replicators and the schemas that are replicated to them.
func (p *Peer) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	log.Debug(ctx, "Received GetAllReplicators request")

	return p.db.GetAllReplicators(ctx)
}

func (p *Peer) loadReplicators(ctx context.Context) error {
	reps, err := p.db.GetAllReplicators(ctx)
	if err != nil {
		return errors.Wrap("failed to get replicators", err)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, rep := range reps {
		for _, schema := range rep.Schemas {
			if pReps, exists := p.replicators[schema]; exists {
				if _, exists := pReps[rep.Info.ID]; exists {
					continue
				}
			} else {
				p.replicators[schema] = make(map[peer.ID]struct{})
			}

			// add to replicators list
			p.replicators[schema][rep.Info.ID] = struct{}{}
		}

		// Add the destination's peer multiaddress in the peerstore.
		// This will be used during connection and stream creation by libp2p.
		p.host.Peerstore().AddAddrs(rep.Info.ID, rep.Info.Addrs, peerstore.PermanentAddrTTL)

		log.Info(ctx, "loaded replicators from datastore", logging.NewKV("Replicator", rep))
	}

	return nil
}

func (p *Peer) loadP2PCollections(ctx context.Context) (map[string]struct{}, error) {
	collections, err := p.db.GetAllP2PCollections(ctx)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return nil, err
	}
	colMap := make(map[string]struct{})
	for _, col := range collections {
		err := p.server.addPubSubTopic(col, true)
		if err != nil {
			return nil, err
		}
		colMap[col] = struct{}{}
	}

	return colMap, nil
}

func (p *Peer) handleDocCreateLog(evt events.Update) error {
	dockey, err := client.NewDocKeyFromString(evt.DocKey)
	if err != nil {
		return NewErrFailedToGetDockey(err)
	}

	// We need to register the document before pushing to the replicators if we want to
	// ensure that we have subscribed to the topic.
	err = p.RegisterNewDocument(p.ctx, dockey, evt.Cid, evt.Block, evt.SchemaID)
	if err != nil {
		return err
	}
	// push to each peer (replicator)
	p.pushLogToReplicators(p.ctx, evt)

	return nil
}

func (p *Peer) handleDocUpdateLog(evt events.Update) error {
	dockey, err := client.NewDocKeyFromString(evt.DocKey)
	if err != nil {
		return NewErrFailedToGetDockey(err)
	}
	log.Debug(
		p.ctx,
		"Preparing pubsub pushLog request from broadcast",
		logging.NewKV("DocKey", dockey),
		logging.NewKV("CID", evt.Cid),
		logging.NewKV("SchemaId", evt.SchemaID))

	body := &pb.PushLogRequest_Body{
		DocKey:   []byte(dockey.String()),
		Cid:      evt.Cid.Bytes(),
		SchemaID: []byte(evt.SchemaID),
		Creator:  p.host.ID().String(),
		Log: &pb.Document_Log{
			Block: evt.Block.RawData(),
		},
	}
	req := &pb.PushLogRequest{
		Body: body,
	}

	// push to each peer (replicator)
	p.pushLogToReplicators(p.ctx, evt)

	if err := p.server.publishLog(p.ctx, evt.DocKey, req); err != nil {
		return NewErrPublishingToDockeyTopic(err, evt.Cid.String(), evt.DocKey)
	}

	if err := p.server.publishLog(p.ctx, evt.SchemaID, req); err != nil {
		return NewErrPublishingToSchemaTopic(err, evt.Cid.String(), evt.SchemaID)
	}

	return nil
}

func (p *Peer) pushLogToReplicators(ctx context.Context, lg events.Update) {
	// push to each peer (replicator)
	peers := make(map[string]struct{})
	for _, peer := range p.ps.ListPeers(lg.DocKey) {
		peers[peer.String()] = struct{}{}
	}
	for _, peer := range p.ps.ListPeers(lg.SchemaID) {
		peers[peer.String()] = struct{}{}
	}

	p.mu.Lock()
	reps, exists := p.replicators[lg.SchemaID]
	p.mu.Unlock()

	if exists {
		for pid := range reps {
			// Don't push if pid is in the list of peers for the topic.
			// It will be handled by the pubsub system.
			if _, ok := peers[pid.String()]; ok {
				continue
			}
			go func(peerID peer.ID) {
				if err := p.server.pushLog(p.ctx, lg, peerID); err != nil {
					log.ErrorE(
						p.ctx,
						"Failed pushing log",
						err,
						logging.NewKV("DocKey", lg.DocKey),
						logging.NewKV("CID", lg.Cid),
						logging.NewKV("PeerID", peerID))
				}
			}(pid)
		}
	}
}

func (p *Peer) setupBlockService() {
	bswapnet := network.NewFromIpfsHost(p.host, p.dht)
	bswap := bitswap.New(p.ctx, bswapnet, p.db.Blockstore())
	p.bserv = blockservice.New(p.db.Blockstore(), bswap)
	p.exch = bswap
}

func (p *Peer) setupDAGService() {
	p.DAGService = dag.NewDAGService(p.bserv)
}

func (p *Peer) newDAGSyncerTxn(txn datastore.Txn) ipld.DAGService {
	return dag.NewDAGService(blockservice.New(txn.DAGstore(), p.exch))
}

// Session returns a session-based NodeGetter.
func (p *Peer) Session(ctx context.Context) ipld.NodeGetter {
	ng := dag.NewSession(ctx, p.DAGService)
	if ng == p.DAGService {
		log.Info(ctx, "DAGService does not support sessions")
	}
	return ng
}

func stopGRPCServer(ctx context.Context, server *grpc.Server) {
	stopped := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(stopped)
	}()
	timer := time.NewTimer(10 * time.Second)
	select {
	case <-timer.C:
		server.Stop()
		log.Info(ctx, "Peer gRPC server was shutdown ungracefully")
	case <-stopped:
		timer.Stop()
	}
}

type EvtReceivedPushLog struct {
	ByPeer   peer.ID
	FromPeer peer.ID
}

type EvtPubSub struct {
	Peer peer.ID
}

// rollbackAddPubSubTopics removes the given topics from the pubsub system.
func (p *Peer) rollbackAddPubSubTopics(cause error, topics ...string) error {
	for _, topic := range topics {
		if err := p.server.removePubSubTopic(topic); err != nil {
			return errors.WithStack(err, errors.NewKV("Cause", cause))
		}
	}
	return cause
}

// rollbackRemovePubSubTopics adds back the given topics from the pubsub system.
func (p *Peer) rollbackRemovePubSubTopics(cause error, topics ...string) error {
	for _, topic := range topics {
		if err := p.server.addPubSubTopic(topic, true); err != nil {
			return errors.WithStack(err, errors.NewKV("Cause", cause))
		}
	}
	return cause
}

// AddP2PCollection adds the given collectionID to the pubsup topics.
//
// It will error if the given collectionID is invalid, in such a case some of the
// changes to the server may still be applied.
//
// WARNING: Calling this on collections with a large number of documents may take a long time to process.
func (p *Peer) AddP2PCollection(
	ctx context.Context,
	collectionID string,
) error {
	log.Debug(ctx, "Received AddP2PCollections request")

	txn, err := p.db.NewTxn(p.ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(p.ctx)
	store := p.db.WithTxn(txn)

	// first let's make sure the collections actually exists
	col, err := store.GetCollectionBySchemaID(p.ctx, collectionID)
	if err != nil {
		return err
	}

	// Ensure we can add all the collections to the store on the transaction
	// before adding to topics.
	err = store.AddP2PCollection(p.ctx, col.SchemaID())
	if err != nil {
		return err
	}

	// Add pubsub topics and remove them if we get an error.
	err = p.server.addPubSubTopic(col.SchemaID(), true)
	if err != nil {
		return p.rollbackAddPubSubTopics(err, col.SchemaID())
	}

	// After adding the collection topics, we remove the collections' documents
	// from the pubsub topics to avoid receiving duplicate events.
	removedTopics := []string{}
	keyChan, err := col.GetAllDocKeys(p.ctx)
	if err != nil {
		return err
	}
	for key := range keyChan {
		err := p.server.removePubSubTopic(key.Key.String())
		if err != nil {
			return p.rollbackRemovePubSubTopics(err, removedTopics...)
		}
		removedTopics = append(removedTopics, key.Key.String())
	}

	if err = txn.Commit(p.ctx); err != nil {
		err = p.rollbackRemovePubSubTopics(err, removedTopics...)
		return p.rollbackAddPubSubTopics(err, col.SchemaID())
	}

	return nil
}

// RemoveP2PCollection removes the given collectionID from the pubsup topics.
//
// It will error if the given collectionID is invalid, in such a case some of the
// changes to the server may still be applied.
//
// WARNING: Calling this on collections with a large number of documents may take a long time to process.
func (p *Peer) RemoveP2PCollection(
	ctx context.Context,
	collectionID string,
) error {
	log.Debug(ctx, "Received RemoveP2PCollections request")

	txn, err := p.db.NewTxn(p.ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(p.ctx)
	store := p.db.WithTxn(txn)

	// first let's make sure the collections actually exists
	col, err := store.GetCollectionBySchemaID(p.ctx, collectionID)
	if err != nil {
		return err
	}

	// Ensure we can remove all the collections to the store on the transaction
	// before adding to topics.
	err = store.RemoveP2PCollection(p.ctx, col.SchemaID())
	if err != nil {
		return err
	}

	// Remove pubsub topics and add them back if we get an error.
	err = p.server.removePubSubTopic(col.SchemaID())
	if err != nil {
		return p.rollbackRemovePubSubTopics(err, col.SchemaID())
	}

	// After removing the collection topics, we add back the collections' documents
	// to the pubsub topics.
	addedTopics := []string{}
	keyChan, err := col.GetAllDocKeys(p.ctx)
	if err != nil {
		return err
	}
	for key := range keyChan {
		err := p.server.addPubSubTopic(key.Key.String(), true)
		if err != nil {
			return p.rollbackAddPubSubTopics(err, addedTopics...)
		}
		addedTopics = append(addedTopics, key.Key.String())
	}

	if err = txn.Commit(p.ctx); err != nil {
		err = p.rollbackAddPubSubTopics(err, addedTopics...)
		return p.rollbackRemovePubSubTopics(err, col.SchemaID())
	}

	return nil
}

// GetAllP2PCollections gets all the collectionIDs from the pubsup topics
func (p *Peer) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	log.Debug(ctx, "Received GetAllP2PCollections request")

	txn, err := p.db.NewTxn(p.ctx, false)
	if err != nil {
		return nil, err
	}
	store := p.db.WithTxn(txn)

	collections, err := store.GetAllP2PCollections(p.ctx)
	if err != nil {
		txn.Discard(p.ctx)
		return nil, err
	}

	return collections, txn.Commit(p.ctx)
}
