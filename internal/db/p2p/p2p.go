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
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
	"github.com/sourcenetwork/defradb/internal/db/permission"
	"github.com/sourcenetwork/defradb/internal/telemetry"
)

var (
	log    = corelog.NewLogger("p2p")
	tracer = telemetry.NewTracer()
)

const networkRequestTimeout = 10 * time.Second

// DB hold the database related methods that are required by P2P.
type DB interface {
	// NewTxn returns a new transaction on the root store that may be managed externally.
	NewTxn(ctx context.Context, readOnly bool) (client.Txn, error)
	// GetNodeIndentityToken returns an identity token for the given audience.
	GetNodeIdentityToken(ctx context.Context, audience immutable.Option[string]) ([]byte, error)
	// GetCollections returns all collections and their descriptions matching the given options
	// that currently exist within this [Store].
	GetCollections(ctx context.Context, options client.CollectionFetchOptions) ([]client.Collection, error)
	// Merge initiates a merge of the DAG and caches the resulting values into the datastore.
	Merge(ctx context.Context, evt event.Merge) error
	// Events returns the event bus for the database.
	Events() event.Bus
	// RetryIntervals returns the replicator retry configuration.
	RetryIntervals() []time.Duration
	// DocumentACP returns the DocumentACP implementation configured on the database.
	DocumentACP() immutable.Option[dac.DocumentACP]
}

type P2P struct {
	identityProtocol   *protocol.IdentityProtocol
	replicatorProtocol *protocol.ReplicatorProtocol

	ctx  context.Context
	db   DB
	host client.Host

	// replicators is a map from collection CollectionID => peerId
	replicators map[string]map[string]client.PeerInfo
	repMu       sync.Mutex

	peerIdentities map[string]identity.Identity
	piMu           sync.RWMutex

	// The intervals at which to retry replicator failures.
	// For example, this can define an exponential backoff strategy.
	retryIntervals   []time.Duration
	handleRetryMutex sync.Mutex
}

// New returns a new configured P2P instance.
func New(ctx context.Context, db DB, host client.Host) (*P2P, error) {
	p := P2P{
		ctx:              ctx,
		db:               db,
		host:             host,
		identityProtocol: protocol.NewIdentityProtocol(host, db.GetNodeIdentityToken),
		replicators:      make(map[string]map[string]client.PeerInfo),
		peerIdentities:   make(map[string]identity.Identity),
		retryIntervals:   db.RetryIntervals(),
	}
	p.replicatorProtocol = protocol.NewReplicatorProtocol(host, p.processPushlogRequest, p.handleReplicatorFailure)

	host.SetBlockAccessFunc(p.hasAccess)

	err := p.host.AddPubSubTopic(docSyncTopic, true, p.docSyncMessageHandler)
	if err != nil {
		return nil, err
	}

	go p.handleReplicatorRetries(ctx)
	err = p.loadAndPublishReplicators(ctx)
	if err != nil {
		return nil, err
	}
	err = p.loadAndPublishP2PCollections(ctx)
	if err != nil {
		return nil, err
	}
	err = p.loadAndPublishP2PDocuments(ctx)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (p *P2P) PeerInfo() client.PeerInfo {
	return p.host.PeerInfo()
}

// Connect initiates a connection to the peer with the given addresp.
func (p *P2P) Connect(ctx context.Context, info client.PeerInfo) error {
	return p.host.Connect(ctx, info)
}

func (p *P2P) updateReplicators(ctx context.Context, rep client.PeerInfo, collectionIDs map[string]struct{}) {
	if len(collectionIDs) == 0 {
		// remove peer from store
		if err := p.host.Disconnect(ctx, rep.ID); err != nil {
			log.ErrorE("Failed to disconnect from replicator peer", err)
		}
	} else {
		if err := p.host.Connect(ctx, rep); err != nil {
			log.ErrorE("Failed to connect to replicator peer", err)
		}
	}

	// update the cached replicators
	p.repMu.Lock()
	for collectionID, peers := range p.replicators {
		if _, hasID := collectionIDs[collectionID]; hasID {
			p.replicators[collectionID][rep.ID] = rep
			delete(collectionIDs, collectionID)
		} else {
			if _, exists := peers[rep.ID]; exists {
				delete(p.replicators[collectionID], rep.ID)
			}
		}
	}
	for collectionID := range collectionIDs {
		if _, exists := p.replicators[collectionID]; !exists {
			p.replicators[collectionID] = make(map[string]client.PeerInfo)
		}
		p.replicators[collectionID][rep.ID] = rep
	}
	p.repMu.Unlock()
}

// hasAccess checks if the requesting peer has access to the given cid.
//
// This is used as a filter in bitswap to determine if we should send the block to the requesting peer.
func (p *P2P) hasAccess(ctx context.Context, pid string, c cid.Cid) bool {
	if !p.db.DocumentACP().HasValue() {
		return true
	}

	clientTxn, err := p.db.NewTxn(ctx, false)
	if err != nil {
		log.ErrorE("Failed to get new transaction", err)
		return false
	}
	defer clientTxn.Discard(ctx)
	txn := datastore.MustGetFromClientTxn(clientTxn)

	rawblock, err := txn.Blockstore().Get(ctx, c)
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
		ctx,
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

	// If the requesting peer is in the replicators list for that collection, then they have accesp.
	p.repMu.Lock()
	if peerList, ok := p.replicators[cols[0].CollectionID()]; ok {
		_, exists := peerList[pid]
		if exists {
			p.repMu.Unlock()
			return true
		}
	}
	p.repMu.Unlock()

	identFunc := func() immutable.Option[identity.Identity] {
		p.piMu.RLock()
		ident, ok := p.peerIdentities[pid]
		p.piMu.RUnlock()
		if !ok {
			ctx, cancel := context.WithTimeout(ctx, networkRequestTimeout)
			defer cancel()
			resp, err := p.identityProtocol.GetIdentity(ctx, pid)
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
			err = identity.VerifyAuthToken(tokenIdent, p.host.ID())
			if err != nil {
				log.ErrorE("Failed to verify auth token", err)
				return immutable.None[identity.Identity]()
			}
			p.piMu.Lock()
			p.peerIdentities[pid] = ident
			p.piMu.Unlock()
		}
		return immutable.Some(ident)
	}

	peerHasAccess, err := permission.CheckDocAccessWithIdentityFunc(
		ctx,
		identFunc,
		p.db.DocumentACP().Value(),
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
func (p *P2P) trySelfHasAccess(ctx context.Context, block *coreblock.Block, collectionID string) (bool, error) {
	if !p.db.DocumentACP().HasValue() {
		return true, nil
	}

	clientTxn, err := p.db.NewTxn(ctx, false)
	if err != nil {
		return false, err
	}
	defer clientTxn.Discard(ctx)

	cols, err := clientTxn.GetCollections(
		ctx,
		client.CollectionFetchOptions{
			CollectionID: immutable.Some(collectionID),
		},
	)
	if err != nil {
		return false, err
	}
	if len(cols) == 0 {
		return false, client.ErrCollectionNotFound
	}
	ident, err := clientTxn.GetNodeIdentity(ctx)
	if err != nil {
		return false, err
	}
	if !ident.HasValue() {
		return true, nil
	}

	peerHasAccess, err := permission.CheckDocAccessWithIdentityFunc(
		ctx,
		func() immutable.Option[identity.Identity] {
			return immutable.Some(identity.FromDID(ident.Value().DID))
		},
		p.db.DocumentACP().Value(),
		cols[0], // For now we assume there is only one collection.
		acpTypes.DocumentReadPerm,
		string(block.Delta.GetDocID()),
	)
	if err != nil {
		return false, err
	}

	return peerHasAccess, nil
}

// pubSubMessageHandler handles incoming PushLog messages from the pubsub network.
func (p *P2P) pubSubMessageHandler(from string, topic string, msg []byte) ([]byte, error) {
	log.Info("Received new pubsub message",
		corelog.String("PeerID", p.host.ID()),
		corelog.Any("SenderId", from),
		corelog.String("Topic", topic))

	req := &protocol.PushLogRequest{}
	if err := cbor.Unmarshal(msg, req); err != nil {
		return nil, err
	}
	req.SenderID = from

	if err := p.processPushlogRequest(p.ctx, req, false); err != nil {
		return nil, errors.Wrap(fmt.Sprintf("Failed to process pushlog request %s", topic), err)
	}

	return nil, nil
}

// processPushlogRequest processes a push log request
func (p *P2P) processPushlogRequest(
	ctx context.Context,
	req *protocol.PushLogRequest,
	isReplicator bool,
) error {
	block, err := coreblock.GetFromBytes(req.Block)
	if err != nil {
		return err
	}

	// No need to check access if the message is for replication as the node sending
	// will have done so deliberately.
	if !isReplicator {
		mightHaveAccess, err := p.trySelfHasAccess(ctx, block, req.CollectionID)
		if err != nil {
			return err
		}
		if !mightHaveAccess {
			// If we know we don't have access, we can skip the rest of the processing.
			return nil
		}
	}

	err = syncDAG(ctx, p.host.BlockService(), block)
	if err != nil {
		return err
	}

	headCID, err := cid.Cast(req.CID)
	if err != nil {
		return err
	}

	go func() {
		evt := event.Merge{
			DocID:        req.DocID,
			ByPeer:       req.SenderID,
			FromPeer:     req.Creator,
			Cid:          headCID,
			CollectionID: req.CollectionID,
		}
		err := p.db.Merge(ctx, evt)
		if err != nil {
			log.ErrorContextE(
				ctx,
				"Failed to execute merge",
				err,
				corelog.Any("Event", evt))
		}
	}()

	return nil
}

func (p *P2P) SendUpdate(evt event.Update) error {
	// push to each peer (replicator)
	p.pushLogToReplicators(evt)

	// Retries are for replicators only and should not pollute the pubsub network.
	if !evt.IsRetry {
		req := &protocol.PushLogRequest{
			DocID:        evt.DocID,
			CID:          evt.Cid.Bytes(),
			CollectionID: evt.CollectionID,
			Creator:      p.host.ID(),
			Block:        evt.Block,
		}

		b, err := cbor.Marshal(req)
		if err != nil {
			return err
		}

		if evt.DocID != "" {
			if err := p.host.PublishToTopicAsync(p.ctx, evt.DocID, b); err != nil {
				return NewErrPublishingToDocIDTopic(err, evt.Cid.String(), evt.DocID)
			}
		}

		if err := p.host.PublishToTopicAsync(p.ctx, evt.CollectionID, b); err != nil {
			return NewErrPublishingToSchemaTopic(err, evt.Cid.String(), evt.CollectionID)
		}
	}

	return nil
}
