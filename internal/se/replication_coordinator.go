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

const (
	// retryLoopInterval is the interval at which the retry handler checks for
	// SE artifacts that are due for a retry. Same as document replicator.
	retryLoopInterval = 2 * time.Second
)

var log = corelog.NewLogger("defra.se.replication")

type P2P interface {
	GetReplicatorsIDs(collectionID string) []string
	Host() client.Host
}

// ReplicationCoordinator manages SE artifact replication and storage
type ReplicationCoordinator struct {
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

type proto[Req, Rep any] interface {
	SendRequest(context.Context, Req, string, bool) (Rep, error)
}

// DB interface required by ReplicationCoordinator
type DB interface {
	Rootstore() corekv.TxnStore
	Events() event.Bus
	MaxTxnRetries() int
	GetCollections(context.Context, client.CollectionFetchOptions) ([]client.Collection, error)
}

// SERetryInfo stores retry information for failed SE replications
type SERetryInfo struct {
	DocID        string
	CollectionID string
	FieldNames   []string
	NextRetry    time.Time
	NumRetries   int
	Retrying     bool
	PublicKey    string // Hex-encoded public key for identity reconstruction
	KeyType      string // Key type (secp256k1, ed25519, etc.)
}

// seStoreProcessor implements CommProcessor for SE artifact storage
type seStoreProcessor struct {
	coordinator *ReplicationCoordinator
}

func (proc *seStoreProcessor) ProcessRequest(
	ctx context.Context,
	req PushSEArtifactsRequest,
	isReplicator bool,
) (PushSEArtifactsReply, error) {
	return PushSEArtifactsReply{}, proc.coordinator.processPushSEArtifactsRequest(ctx, &req, isReplicator)
}

// seQueryProcessor implements CommProcessor for SE artifact queries
type seQueryProcessor struct {
	coordinator *ReplicationCoordinator
}

func (proc *seQueryProcessor) ProcessRequest(
	ctx context.Context,
	req QuerySEArtifactsRequest,
	isReplicator bool,
) (QuerySEArtifactsReply, error) {
	return proc.coordinator.processQuerySEArtifactsRequest(ctx, &req, isReplicator)
}

// NewReplicationCoordinator creates a new coordinator
func NewReplicationCoordinator(db DB, p2p P2P, encKey []byte) (*ReplicationCoordinator, error) {
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
) (*ReplicationCoordinator, error) {
	ctx, cancel := context.WithCancel(context.Background())

	rc := &ReplicationCoordinator{
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

func (rc *ReplicationCoordinator) Close() {
	rc.cancel()
	rc.eventBus.Unsubscribe(rc.sub)
}

// reconstructIdentity reconstructs an Identity from stored public key information
func (rc *ReplicationCoordinator) reconstructIdentity(
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

// defaultRetryIntervals generates retry intervals based on max retries
func defaultRetryIntervals(maxRetries int) []time.Duration {
	intervals := make([]time.Duration, maxRetries)
	for i := range maxRetries {
		// Exponential backoff: 2s, 4s, 8s, 16s...
		intervals[i] = time.Second * time.Duration(2<<i)
	}
	return intervals
}

// processUpdateEvents handles updates to SE artifacts
func (rc *ReplicationCoordinator) processEvents() {
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

func (rc *ReplicationCoordinator) handleQuerySEArtifactsEvent(evt RequestSEArtifactsEvent) {
	grpcQueries := make([]SEFieldQuery, len(evt.Queries))
	for i, q := range evt.Queries {
		grpcQueries[i] = SEFieldQuery(q)
	}

	grpcReq := QuerySEArtifactsRequest{
		CollectionID: evt.CollectionID,
		Queries:      grpcQueries,
	}

	docIDSet := make(map[string]struct{})
	var queryErr error

	peerIDs := rc.p2p.GetReplicatorsIDs(evt.CollectionID)

	if len(peerIDs) == 0 {
		evt.Response <- SEArtifactsResult{
			DocIDs: []string{},
			Error:  nil,
		}
		return
	}

	for _, pid := range peerIDs {
		reply, err := rc.querySEProto.SendRequest(context.Background(), grpcReq, pid, false)
		if err != nil {
			log.ErrorE(
				"Failed querying SE artifacts from replicator",
				err,
				corelog.String("CollectionID", evt.CollectionID),
				corelog.Any("PeerID", pid))
			queryErr = err
			continue
		}

		for _, docID := range reply.DocIDs {
			docIDSet[docID] = struct{}{}
		}
		break
	}

	docIDs := make([]string, 0, len(docIDSet))
	for docID := range docIDSet {
		docIDs = append(docIDs, docID)
	}

	evt.Response <- SEArtifactsResult{
		DocIDs: docIDs,
		Error:  queryErr,
	}
}

// handleReplicationFailure stores failed SE replication attempt for retry
func (rc *ReplicationCoordinator) handleReplicationFailure(
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
func (rc *ReplicationCoordinator) handleUpdateEvent(ctx context.Context, evt event.Update) error {
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

func (rc *ReplicationCoordinator) generateArtifactsAndPushToReplicators(
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
		_, err = rc.storeSEProto.SendRequest(ctx, req, pid, false)
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

// retrySEReplicators periodically processes failed SE replications
func (rc *ReplicationCoordinator) retrySEReplicators(ctx context.Context) {
	ticker := time.NewTicker(retryLoopInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rc.processSERetries(ctx)
		}
	}
}

// processSERetries checks for due retries and processes them
func (rc *ReplicationCoordinator) processSERetries(ctx context.Context) {
	ps := datastore.PeerstoreFrom(rc.db.Rootstore())
	iter, err := ps.Iterator(ctx, corekv.IterOptions{
		Prefix: keys.NewPeerstoreSERetry("", "", "").Bytes(),
	})
	if err != nil {
		log.ErrorContextE(ctx, "Failed to iterate SE retry keys", err)
		return
	}

	now := time.Now()
	for {
		hasNext, err := iter.Next()
		if err != nil {
			log.ErrorContextE(ctx, "Failed to get next SE retry key", err)
			break
		}
		if !hasNext {
			break
		}

		value, err := iter.Value()
		if err != nil {
			log.ErrorContextE(ctx, "Failed to get SE retry value", err)
			continue
		}

		retryInfo := SERetryInfo{}
		err = cbor.Unmarshal(value, &retryInfo)
		if err != nil {
			log.ErrorContextE(ctx, "Failed to unmarshal SE retry info", err)
			continue
		}

		// Check if retry is due and not already in progress
		if now.After(retryInfo.NextRetry) && !retryInfo.Retrying {
			key, err := keys.NewPeerstoreSERetryFromString(string(iter.Key()))
			if err != nil {
				log.ErrorContextE(ctx, "Failed to parse SE retry key", err)
				continue
			}

			retryInfo.Retrying = true
			retryInfo.NumRetries++
			b, err := cbor.Marshal(retryInfo)
			if err != nil {
				log.ErrorContextE(ctx, "Failed to marshal SE retry info", err)
				continue
			}
			ps := datastore.PeerstoreFrom(rc.db.Rootstore())
			if err := ps.Set(ctx, iter.Key(), b); err != nil {
				log.ErrorContextE(ctx, "Failed to update SE retry info", err)
				continue
			}

			go rc.retrySEArtifacts(ctx, key.PeerID, retryInfo)
		}
	}

	err = iter.Close()
	if err != nil {
		log.ErrorContextE(ctx, "Failed to close SE retry iterator", err)
	}
}

// retrySEArtifacts attempts to retry SE artifact replication for a document
//
// Note: This function relies on the SE artifact generation phase to re-generate
// artifacts from the document's field values. We don't store SE artifacts locally
// on the producer node - they are only stored on replicator nodes.
func (rc *ReplicationCoordinator) retrySEArtifacts(ctx context.Context, peerID string, retryInfo SERetryInfo) {
	log.InfoContext(ctx, "Retrying SE replicator", corelog.String("PeerID", peerID))

	identity, err := rc.reconstructIdentity(retryInfo.PublicKey, retryInfo.KeyType)
	if err != nil {
		log.ErrorContextE(ctx, "Failed to reconstruct identity from stored data", err,
			corelog.String("DocID", retryInfo.DocID))
	} else if identity.HasValue() {
		ctx = acpIdentity.WithContext(ctx, identity)
	}

	err = rc.generateArtifactsAndPushToReplicators(ctx, retryInfo.DocID,
		retryInfo.CollectionID, retryInfo.FieldNames, identity, true)
	if err != nil {
		log.ErrorContextE(ctx, "Failed to generate and push SE artifacts for retry", err,
			corelog.String("DocID", retryInfo.DocID))
	}

	rc.updateRetryStatus(ctx, peerID, retryInfo, err == nil)
}

// updateRetryStatus updates the retry status after an attempt
func (rc *ReplicationCoordinator) updateRetryStatus(
	ctx context.Context,
	peerID string,
	retryInfo SERetryInfo,
	success bool,
) {
	retryKey := keys.NewPeerstoreSERetry(peerID, retryInfo.CollectionID, retryInfo.DocID)

	if success {
		ps := datastore.PeerstoreFrom(rc.db.Rootstore())
		if err := ps.Delete(ctx, retryKey.Bytes()); err != nil {
			log.ErrorContextE(ctx, "Failed to delete SE retry entry", err)
		}
	} else {
		if retryInfo.NumRetries >= len(rc.retryIntervals) {
			retryInfo.NextRetry = time.Now().Add(rc.retryIntervals[len(rc.retryIntervals)-1])
		} else {
			retryInfo.NextRetry = time.Now().Add(rc.retryIntervals[retryInfo.NumRetries])
		}
		retryInfo.Retrying = false

		b, err := cbor.Marshal(retryInfo)
		if err != nil {
			log.ErrorContextE(ctx, "Failed to marshal SE retry info", err)
			return
		}
		ps := datastore.PeerstoreFrom(rc.db.Rootstore())
		if err := ps.Set(ctx, retryKey.Bytes(), b); err != nil {
			log.ErrorContextE(ctx, "Failed to update SE retry info", err)
		}
	}
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
func (rc *ReplicationCoordinator) DeleteSEArtifacts(
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
func (rc *ReplicationCoordinator) generateSEArtifacts(
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

func (rc *ReplicationCoordinator) processPushSEArtifactsRequest(
	ctx context.Context,
	req *PushSEArtifactsRequest,
	isReplicator bool,
) error {
	sb := strings.Builder{}
	for i, netArtifact := range req.Artifacts {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(netArtifact.DocID)
	}
	log.InfoContext(ctx, "Handle push SE artifacts",
		corelog.String("DocIDs", sb.String()), corelog.String("Sender", req.SenderID))

	artifacts := make([]secore.Artifact, len(req.Artifacts))
	for i, netArtifact := range req.Artifacts {
		artifacts[i] = secore.Artifact{
			DocID:        netArtifact.DocID,
			IndexID:      netArtifact.IndexID,
			SearchTag:    netArtifact.SearchTag,
			CollectionID: req.CollectionID,
		}
	}

	if err := StoreArtifacts(ctx, datastore.DatastoreFrom(rc.db.Rootstore()), artifacts); err != nil {
		return err
	}

	return nil
}

func (rc *ReplicationCoordinator) processQuerySEArtifactsRequest(
	ctx context.Context,
	req *QuerySEArtifactsRequest,
	isReplicator bool,
) (QuerySEArtifactsReply, error) {
	matchingDocIDs, err := rc.querySEArtifactsFromDatastore(ctx, req)
	if err != nil {
		log.ErrorContextE(ctx, "Failed to query SE artifacts from datastore", err)
		return QuerySEArtifactsReply{}, err
	}

	log.InfoContext(ctx, "Handle SE artifacts query", corelog.String("DocIDs", strings.Join(matchingDocIDs, ", ")),
		corelog.String("Sender", req.SenderID))

	return QuerySEArtifactsReply{
		DocIDs: matchingDocIDs,
	}, nil
}

// querySEArtifactsFromDatastore queries SE artifacts from the local datastore
func (rc *ReplicationCoordinator) querySEArtifactsFromDatastore(
	ctx context.Context,
	req *QuerySEArtifactsRequest,
) ([]string, error) {
	queries := make([]FieldQuery, len(req.Queries))
	for i, q := range req.Queries {
		queries[i] = FieldQuery(q)
	}

	return FetchDocIDs(ctx, datastore.DatastoreFrom(rc.db.Rootstore()), req.CollectionID, queries)
}
