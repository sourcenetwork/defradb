// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package se

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/fxamacker/cbor/v2"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
	"github.com/sourcenetwork/defradb/internal/keys"
	secore "github.com/sourcenetwork/defradb/internal/se/core"
)

var log = corelog.NewLogger("defra.se.replication")

type P2P interface {
	GetReplicatorsIDs(collectionID string) []string
	Host() client.Host
}

// Coordinator manages SE artifact replication and storage
type Coordinator struct {
	db             DB
	eventBus       event.Bus
	sub            event.Subscription
	retryIntervals []time.Duration
	encKey         []byte // Encryption key for SE artifacts
	p2p            P2P
	storeSEProto   proto[PushSEArtifactsRequest, PushSEArtifactsReply]
	querySEProto   proto[QuerySEArtifactsRequest, QuerySEArtifactsReply]

	ctx    context.Context
	cancel context.CancelFunc
}

// DB interface required by ReplicationCoordinator
type DB interface {
	Rootstore() corekv.TxnStore
	Events() event.Bus
	MaxTxnRetries() int
	GetCollections(context.Context, client.CollectionFetchOptions) ([]client.Collection, error)
}

// NewReplicationCoordinator creates a new coordinator
func NewReplicationCoordinator(db DB, p2p P2P, encKey []byte) (*Coordinator, error) {
	rc, err := newReplicationCoordinator(
		db,
		p2p,
		encKey,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	rc.storeSEProto = protocol.NewCommChannel(
		p2p.Host(),
		"rep_se",
		&seStoreProcessor{coordinator: rc},
	)
	rc.querySEProto = protocol.NewCommChannel(
		p2p.Host(),
		"se_query",
		&seQueryProcessor{coordinator: rc},
	)

	return rc, nil
}

func newReplicationCoordinator(
	db DB,
	p2p P2P,
	encKey []byte,
	push proto[PushSEArtifactsRequest, PushSEArtifactsReply],
	query proto[QuerySEArtifactsRequest, QuerySEArtifactsReply],
) (*Coordinator, error) {
	ctx, cancel := context.WithCancel(context.Background())

	rc := &Coordinator{
		db:             db,
		eventBus:       db.Events(),
		retryIntervals: defaultRetryIntervals(db.MaxTxnRetries()),
		encKey:         encKey,
		p2p:            p2p,
		ctx:            ctx,
		cancel:         cancel,
		storeSEProto:   push,
		querySEProto:   query,
	}

	var err error
	rc.sub, err = db.Events().Subscribe(event.UpdateName, QuerySEArtifactsEventName)
	if err != nil {
		return nil, err
	}

	go rc.processEvents()

	go rc.retrySEReplicators(rc.ctx)

	return rc, nil
}

func (rc *Coordinator) Close() {
	rc.cancel()
	rc.eventBus.Unsubscribe(rc.sub)
}

// reconstructIdentity reconstructs an Identity from stored public key information
func (rc *Coordinator) reconstructIdentity(
	publicKey, keyType string,
) (immutable.Option[acpIdentity.Identity], error) {
	if publicKey == "" || keyType == "" {
		return immutable.None[acpIdentity.Identity](), nil
	}

	pubKey, err := crypto.PublicKeyFromString(crypto.KeyType(keyType), publicKey)
	if err != nil {
		return immutable.None[acpIdentity.Identity](), err
	}

	identity, err := acpIdentity.FromPublicKey(pubKey)
	if err != nil {
		return immutable.None[acpIdentity.Identity](), err
	}

	return immutable.Some(identity), nil
}

// processUpdateEvents handles updates to SE artifacts
func (rc *Coordinator) processEvents() {
	for {
		msg, isOpen := <-rc.sub.Message()
		if !isOpen {
			return
		}

		switch evt := msg.Data.(type) {
		case event.Update:
			if err := rc.handleUpdateEvent(context.Background(), evt); err != nil {
				log.ErrorE("Failed to handle SE update event", err,
					corelog.String("DocID", evt.DocID))
			}

		case RequestSEArtifactsEvent:
			go rc.handleQuerySEArtifactsEvent(evt)

		default:
			continue
		}
	}
}

func (rc *Coordinator) handleQuerySEArtifactsEvent(evt RequestSEArtifactsEvent) {
	grpcQueries := make([]SEFieldQuery, len(evt.Queries))
	for i, q := range evt.Queries {
		grpcQueries[i] = SEFieldQuery(q)
	}

	grpcReq := QuerySEArtifactsRequest{
		CollectionID: evt.CollectionID,
		Queries:      grpcQueries,
	}

	peerIDs := rc.p2p.GetReplicatorsIDs(evt.CollectionID)

	if len(peerIDs) == 0 {
		evt.Response <- SEArtifactsResult{
			DocIDs: []string{},
			Error:  nil,
		}
		return
	}

	var err error
	var reply QuerySEArtifactsReply
	for _, pid := range peerIDs {
		reply, err = rc.querySEProto.SendRequest(context.Background(), grpcReq, pid)
		if err != nil {
			log.ErrorE(
				"Failed querying SE artifacts from replicator",
				err,
				corelog.String("CollectionID", evt.CollectionID),
				corelog.Any("PeerID", pid))
			continue
		}

		break
	}

	evt.Response <- SEArtifactsResult{
		DocIDs: reply.DocIDs,
		Error:  err,
	}
}

// handleReplicationFailure stores failed SE replication attempt for retry
func (rc *Coordinator) handleReplicationFailure(
	ctx context.Context,
	docID, collectionID, peerID string,
	fieldNames []string,
	identity immutable.Option[acpIdentity.Identity],
) error {
	retryKey := keys.NewPeerstoreSERetry(peerID, collectionID, docID)

	var publicKey string
	var keyType string
	if identity.HasValue() {
		identity := identity.Value()
		if pubKey := identity.PublicKey(); pubKey != nil {
			publicKey = hex.EncodeToString(pubKey.Raw())
			keyType = string(pubKey.Type())
		}
	}

	retryInfo := SERetryInfo{
		DocID:        docID,
		CollectionID: collectionID,
		FieldNames:   fieldNames,
		NextRetry:    time.Now().Add(rc.retryIntervals[0]),
		NumRetries:   0,
		PublicKey:    publicKey,
		KeyType:      keyType,
	}

	b, err := cbor.Marshal(retryInfo)
	if err != nil {
		return err
	}

	ps := datastore.PeerstoreFrom(rc.db.Rootstore())
	return ps.Set(ctx, retryKey.Bytes(), b)
}

// handleUpdateEvent processes SE update events and stores artifacts
func (rc *Coordinator) handleUpdateEvent(ctx context.Context, evt event.Update) error {
	// If this is a retry, we don't need to generate artifacts
	if evt.IsRetry {
		return nil
	}

	block, err := coreblock.GetFromBytes(evt.Block)
	if err != nil {
		return fmt.Errorf("failed to deserialize block: %w", err)
	}

	if !block.Delta.IsComposite() {
		return nil
	}

	updatedFields := []string{}
	for _, link := range block.Links {
		if link.Name != "" && link.Name != "_head" {
			updatedFields = append(updatedFields, link.Name)
		}
	}

	if evt.Identity.HasValue() {
		ctx = acpIdentity.WithContext(ctx, evt.Identity)
	}

	return rc.generateArtifactsAndPushToReplicators(ctx, evt.DocID, evt.CollectionID, updatedFields, evt.Identity, false)
}

func (rc *Coordinator) generateArtifactsAndPushToReplicators(
	ctx context.Context,
	docID, collectionID string,
	fields []string,
	identity immutable.Option[acpIdentity.Identity],
	isRetry bool,
) error {
	artifacts, err := rc.generateSEArtifacts(ctx, docID, collectionID, fields)
	if err != nil {
		return fmt.Errorf("failed to generate SE artifacts: %w", err)
	}
	if len(artifacts) == 0 {
		return nil
	}

	protoArtifacts := make([]SEArtifact, len(artifacts))
	for i, artifact := range artifacts {
		protoArtifacts[i] = SEArtifact{
			DocID:     artifact.DocID,
			IndexID:   artifact.IndexID,
			SearchTag: artifact.SearchTag,
		}
	}

	req := PushSEArtifactsRequest{
		CollectionID: collectionID,
		Artifacts:    protoArtifacts,
	}

	peerIDs := rc.p2p.GetReplicatorsIDs(collectionID)
	for _, pid := range peerIDs {
		_, err = rc.storeSEProto.SendRequest(ctx, req, pid)
		if err != nil {
			if isRetry {
				return err
			}
			handleErr := rc.handleReplicationFailure(ctx, docID, collectionID, pid, fields, identity)
			if handleErr != nil {
				return errors.Join(err, handleErr)
			}
		}
	}

	return nil
}

// DeleteSEArtifacts removes SE artifacts from the datastore.
//
// Parameters:
//   - searchTags: If provided, only delete artifacts with these specific search tags.
//     If empty/nil, delete all artifacts for the given document/index combination.
//
// This is typically called when:
//   - A document is deleted (searchTags is empty)
//   - A field value changes (searchTags contains the old search tags to remove)
func (rc *Coordinator) DeleteSEArtifacts(
	ctx context.Context,
	collectionID, indexID, docID string,
	searchTags [][]byte,
) error {
	ds := datastore.DatastoreFrom(rc.db.Rootstore())

	if len(searchTags) > 0 {
		for _, tag := range searchTags {
			key := keys.DatastoreSE{
				CollectionID: collectionID,
				IndexID:      indexID,
				SearchTag:    tag,
				DocID:        docID,
			}
			if err := ds.Delete(ctx, key.Bytes()); err != nil {
				return err
			}
		}
		return nil
	}

	prefix := keys.DatastoreSE{
		CollectionID: collectionID,
		IndexID:      indexID,
	}.Bytes()

	keysToDelete, err := datastore.FetchKeysForPrefix(ctx, prefix, ds)
	if err != nil {
		return err
	}

	for _, key := range keysToDelete {
		keyStr := string(key)
		if strings.HasSuffix(keyStr, "/"+docID) {
			if err := ds.Delete(ctx, key); err != nil {
				return err
			}
		}
	}

	return nil
}

// generateSEArtifacts regenerates SE artifacts for specified fields
//
// This method uses the extracted GenerateArtifacts function to recreate artifacts
// needed for retry.
func (rc *Coordinator) generateSEArtifacts(
	ctx context.Context,
	docID, collectionID string,
	fieldNames []string,
) ([]secore.Artifact, error) {
	cols, err := rc.db.GetCollections(ctx, client.CollectionFetchOptions{
		CollectionID: immutable.Some(collectionID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("collection not found: %s", collectionID)
	}

	col := cols[0]
	docIDType, err := client.NewDocIDFromString(docID)
	if err != nil {
		return nil, fmt.Errorf("invalid document ID: %w", err)
	}

	doc, err := col.Get(ctx, docIDType, false)
	if err != nil {
		if errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return GenerateDocArtifacts(ctx, col, doc, fieldNames, rc.encKey)
}
