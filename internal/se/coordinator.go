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
	"time"

	"github.com/fxamacker/cbor/v2"

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

// DB defines the database operations needed by the SE coordinator
type DB interface {
	NewTxn(ctx context.Context, readOnly bool) (client.Txn, error)
	MaxTxnRetries() int
	GetCollections(context.Context, client.CollectionFetchOptions) ([]client.Collection, error)
	Events() event.Bus
}

type P2P interface {
	GetReplicatorsIDs(collectionID string) []string
	Host() client.Host
}

// Coordinator manages SE artifact replication and storage
type Coordinator struct {
	retryIntervals []time.Duration
	encKey         []byte // Encryption key for SE artifacts
	p2p            P2P
	db             DB
	storeSEProto   protocol.CommChannel[PushSEArtifactsRequest, PushSEArtifactsReply]
	querySEProto   protocol.CommChannel[QuerySEArtifactsRequest, QuerySEArtifactsReply]

	ctx    context.Context
	cancel context.CancelFunc
}

// NewCoordinator creates a new coordinator
func NewCoordinator(p2p P2P, db DB, encKey []byte) (*Coordinator, error) {
	rc, err := NewCoordinatorConfigure(
		p2p,
		db,
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

func NewCoordinatorConfigure(
	p2p P2P,
	db DB,
	encKey []byte,
	push protocol.CommChannel[PushSEArtifactsRequest, PushSEArtifactsReply],
	query protocol.CommChannel[QuerySEArtifactsRequest, QuerySEArtifactsReply],
) (*Coordinator, error) {
	ctx, cancel := context.WithCancel(context.Background())

	rc := &Coordinator{
		retryIntervals: defaultRetryIntervals(db.MaxTxnRetries()),
		encKey:         encKey,
		p2p:            p2p,
		db:             db,
		ctx:            ctx,
		cancel:         cancel,
		storeSEProto:   push,
		querySEProto:   query,
	}

	go rc.retrySEReplicators(rc.ctx)

	return rc, nil
}

func (rc *Coordinator) Close() {
	rc.cancel()
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

// FieldValueQuery represents a field value to query for SE artifacts.
type FieldValueQuery struct {
	FieldName string
	IndexDesc client.EncryptedIndexDescription
	Value     client.NormalValue
}

// QueryDocIDsByValues queries SE artifacts from replicators based on field values.
// It generates search tags from the values and queries replicators for matching documents.
func (rc *Coordinator) QueryDocIDsByValues(
	ctx context.Context,
	collectionID string,
	fieldValues []FieldValueQuery,
) ([]string, error) {
	queries := make([]FieldQuery, 0, len(fieldValues))

	for _, fv := range fieldValues {
		// Generate search tag
		artifact, err := generateFieldArtifact(
			ctx,
			collectionID,
			"", // docID not needed for search tag generation
			fv.IndexDesc,
			fv.Value,
			rc.encKey,
		)
		if err != nil {
			return nil, err
		}

		queries = append(queries, FieldQuery{
			FieldName: fv.FieldName,
			IndexID:   fv.FieldName,
			SearchTag: artifact.SearchTag,
		})
	}

	return rc.QuerySEArtifacts(ctx, collectionID, queries)
}

// QuerySEArtifacts queries SE artifacts from replicators and returns matching document IDs.
// This is called directly by the planner when executing SE queries.
func (rc *Coordinator) QuerySEArtifacts(
	ctx context.Context,
	collectionID string,
	queries []FieldQuery,
) ([]string, error) {
	grpcQueries := make([]SEFieldQuery, len(queries))
	for i, q := range queries {
		grpcQueries[i] = SEFieldQuery(q)
	}

	grpcReq := QuerySEArtifactsRequest{
		CollectionID: collectionID,
		Queries:      grpcQueries,
	}

	peerIDs := rc.p2p.GetReplicatorsIDs(collectionID)

	if len(peerIDs) == 0 {
		return []string{}, nil
	}

	var err error
	var reply QuerySEArtifactsReply
	for _, pid := range peerIDs {
		reply, err = rc.querySEProto.SendRequest(ctx, grpcReq, pid)
		if err != nil {
			// Log the error and try the next peer
			log.ErrorContextE(ctx,
				"Failed querying SE artifacts from replicator",
				err,
				corelog.String("CollectionID", collectionID),
				corelog.Any("PeerID", pid))
		} else {
			// if successful, no need to try other peers
			break
		}
	}

	// If all peers failed, return the last error
	if err != nil {
		return nil, err
	}

	return reply.DocIDs, nil
}

// handleReplicationFailure stores failed SE replication attempt for retry
func (rc *Coordinator) handleReplicationFailure(
	ctx context.Context,
	docID, collectionID, peerID string,
	fieldNames []string,
	identity immutable.Option[acpIdentity.Identity],
) error {
	clientTxn, err := rc.db.NewTxn(ctx, true)
	if err != nil {
		return err
	}
	defer clientTxn.Discard(ctx)
	txn := datastore.MustGetFromClientTxn(clientTxn)
	ctx = datastore.CtxSetTxn(ctx, txn)

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

	err = txn.Peerstore().Set(ctx, retryKey.Bytes(), b)
	if err != nil {
		return nil
	}

	return txn.Commit(ctx)
}

// HandlePushToReplicators processes document update events and generates SE artifacts.
// This implements the PushToReplicatorsHandler interface for P2P.
func (rc *Coordinator) HandlePushToReplicators(ctx context.Context, evt event.Update) error {
	// If this is a retry, we don't need to generate artifacts
	if evt.IsRetry {
		return nil
	}

	block, err := coreblock.GetFromBytes(evt.Block)
	if err != nil {
		return NewErrFailedToDeserializeBlock(err)
	}

	if !block.Delta.IsComposite() {
		return nil
	}

	updatedFields := []string{}
	for _, link := range block.Links {
		updatedFields = append(updatedFields, link.Name)
	}

	if evt.Identity.HasValue() {
		ctx = acpIdentity.WithContext(ctx, evt.Identity)
	}

	return rc.generateArtifactsAndPushToReplicators(ctx, evt.DocID, evt.CollectionID, updatedFields, evt.Identity, false)
}

// generateArtifactsAndPushToReplicators generates SE artifacts and pushes them to replicators.
// This is called by the P2P layer when document updates occur.
func (rc *Coordinator) generateArtifactsAndPushToReplicators(
	ctx context.Context,
	docID, collectionID string,
	fields []string,
	identity immutable.Option[acpIdentity.Identity],
	isRetry bool,
) error {
	artifacts, err := rc.generateSEArtifacts(ctx, docID, collectionID, fields)
	if err != nil {
		return NewErrFailedToGenerateSEArtifacts(err)
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
		return nil, NewErrFailedToGetCollection(err)
	}
	if len(cols) == 0 {
		return nil, NewErrCollectionNotFound(collectionID)
	}

	col := cols[0]
	docIDType, err := client.NewDocIDFromString(docID)
	if err != nil {
		return nil, NewErrInvalidDocumentID(err)
	}

	doc, err := col.Get(ctx, docIDType, false)
	if err != nil {
		if errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) {
			return nil, nil
		}
		return nil, NewErrFailedToGetDocument(err)
	}

	return generateDocArtifacts(ctx, col, doc, fieldNames, rc.encKey)
}
