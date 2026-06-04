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
	"time"

	"github.com/fxamacker/cbor/v2"

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/keys"
	secore "github.com/sourcenetwork/defradb/internal/se/core"
)

var log = corelog.NewLogger("se")

// DB defines the database operations needed by the SE coordinator
type DB interface {
	MaxTxnRetries() int
	GetCollections(context.Context, ...options.Enumerable[options.GetCollectionsOptions]) ([]client.Collection, error)
	Events() event.Bus
	Multistore() *datastore.Multistore
}

type P2P interface {
	GetReplicatorsIDs(collectionID string) []string
}

// Coordinator manages SE artifact replication and storage
type Coordinator struct {
	retryIntervals []time.Duration
	encKey         []byte // Encryption key for SE artifacts
	p2p            P2P
	db             DB
	storeSEProto   protocol.CommChannel[PushSEArtifactsRequest, PushSEArtifactsReply]
	querySEProto   protocol.CommChannel[QuerySEArtifactsRequest, QuerySEArtifactsReply]

	ctx          context.Context
	cancel       context.CancelFunc
	nodeIdentity immutable.Option[acpIdentity.Identity]
}

// NewCoordinator creates a new coordinator
func NewCoordinator(
	p2p P2P,
	host client.Host,
	db DB,
	encKey []byte,
	nodeIdentity immutable.Option[acpIdentity.Identity],
) (*Coordinator, error) {
	coordinator, err := NewCoordinatorConfigure(
		p2p,
		db,
		encKey,
		nil,
		nil,
		nodeIdentity,
	)
	if err != nil {
		return nil, err
	}

	coordinator.storeSEProto = protocol.NewCommChannel(
		host,
		"rep_se",
		&seStoreProcessor{coordinator: coordinator},
	)
	coordinator.querySEProto = protocol.NewCommChannel(
		host,
		"se_query",
		&seQueryProcessor{coordinator: coordinator},
	)

	return coordinator, nil
}

func NewCoordinatorConfigure(
	p2p P2P,
	db DB,
	encKey []byte,
	push protocol.CommChannel[PushSEArtifactsRequest, PushSEArtifactsReply],
	query protocol.CommChannel[QuerySEArtifactsRequest, QuerySEArtifactsReply],
	nodeIdentity immutable.Option[acpIdentity.Identity],
) (*Coordinator, error) {
	ctx, cancel := context.WithCancel(context.Background())

	coordinator := &Coordinator{
		retryIntervals: defaultRetryIntervals(db.MaxTxnRetries()),
		encKey:         encKey,
		p2p:            p2p,
		db:             db,
		ctx:            ctx,
		cancel:         cancel,
		storeSEProto:   push,
		querySEProto:   query,
		nodeIdentity:   nodeIdentity,
	}

	go coordinator.retrySEReplicators(coordinator.ctx)

	return coordinator, nil
}

func (coordinator *Coordinator) Close() {
	coordinator.cancel()
}

// FieldValueQuery represents a field value to query for SE artifacts.
type FieldValueQuery struct {
	FieldName string
	IndexDesc client.EncryptedIndexDescription
	Value     client.NormalValue
}

// QueryDocIDsByValues queries SE artifacts from replicators based on field values.
// It generates search tags from the values and queries replicators for matching documents.
func (coordinator *Coordinator) QueryDocIDsByValues(
	ctx context.Context,
	collectionID string,
	fieldValues []FieldValueQuery,
) ([]string, error) {
	queries := make([]fieldQuery, 0, len(fieldValues))

	for _, fv := range fieldValues {
		// Generate search tag
		artifact, err := generateFieldArtifact(
			collectionID,
			"", // docID not needed for search tag generation
			fv.IndexDesc,
			fv.Value,
			coordinator.nodeIdentity,
			coordinator.encKey,
		)
		if err != nil {
			return nil, err
		}

		queries = append(queries, fieldQuery{
			FieldName: fv.FieldName,
			IndexID:   fv.FieldName,
			SearchTag: artifact.SearchTag,
		})
	}

	return coordinator.QuerySEArtifacts(ctx, collectionID, queries)
}

// QuerySEArtifacts queries SE artifacts from replicators and returns matching document IDs.
// This is called directly by the planner when executing SE queries.
func (coordinator *Coordinator) QuerySEArtifacts(
	ctx context.Context,
	collectionID string,
	queries []fieldQuery,
) ([]string, error) {
	grpcQueries := make([]SEFieldQuery, len(queries))
	for i, q := range queries {
		grpcQueries[i] = SEFieldQuery(q)
	}

	grpcReq := QuerySEArtifactsRequest{
		CollectionID: collectionID,
		Queries:      grpcQueries,
	}

	peerIDs := coordinator.p2p.GetReplicatorsIDs(collectionID)

	if len(peerIDs) == 0 {
		return []string{}, nil
	}

	var err error
	var reply QuerySEArtifactsReply
	for _, pid := range peerIDs {
		reply, err = coordinator.querySEProto.SendRequest(ctx, grpcReq, pid)
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
func (coordinator *Coordinator) handleReplicationFailure(
	ctx context.Context,
	docID, collectionID, peerID string,
	fieldNames []string,
) error {
	log.InfoContext(ctx, "SE replication failed, scheduling retry",
		corelog.String("DocID", docID),
		corelog.String("CollectionID", collectionID))

	retryKey := keys.NewPeerstoreSERetry(peerID, collectionID, docID)

	retryInfo := seRetryInfo{
		DocID:        docID,
		CollectionID: collectionID,
		FieldNames:   fieldNames,
		NextRetry:    time.Now().Add(coordinator.retryIntervals[0]),
		NumRetries:   0,
	}

	b, err := cbor.Marshal(retryInfo)
	if err != nil {
		return err
	}

	return coordinator.db.Multistore().Peerstore().Set(ctx, retryKey.Bytes(), b)
}

// HandlePushToReplicators processes document update events and generates SE artifacts.
// This implements the PushToReplicatorsHandler interface for P2P.
func (coordinator *Coordinator) HandlePushToReplicators(ctx context.Context, evt event.Update) error {
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

	return coordinator.generateArtifactsAndPushToReplicators(ctx, evt.DocID, evt.CollectionID, updatedFields, false)
}

// generateArtifactsAndPushToReplicators generates SE artifacts and pushes them to replicators.
// This is called by the P2P layer when document updates occur.
func (coordinator *Coordinator) generateArtifactsAndPushToReplicators(
	ctx context.Context,
	docID, collectionID string,
	fields []string,
	isRetry bool,
) error {
	artifacts, err := coordinator.generateSEArtifacts(ctx, docID, collectionID, fields)
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

	peerIDs := coordinator.p2p.GetReplicatorsIDs(collectionID)
	for _, pid := range peerIDs {
		_, err = coordinator.storeSEProto.SendRequest(ctx, req, pid)
		if err != nil {
			if isRetry {
				return err
			}
			handleErr := coordinator.handleReplicationFailure(ctx, docID, collectionID, pid, fields)
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
func (coordinator *Coordinator) generateSEArtifacts(
	ctx context.Context,
	docID, collectionID string,
	fieldNames []string,
) ([]secore.Artifact, error) {
	cols, err := coordinator.db.GetCollections(ctx, options.GetCollections().SetCollectionID(collectionID))
	if err != nil {
		return nil, err
	}
	if len(cols) == 0 {
		return nil, NewErrCollectionNotFound(collectionID)
	}

	col := cols[0]
	docIDType, err := client.NewDocIDFromString(docID)
	if err != nil {
		return nil, err
	}

	if coordinator.nodeIdentity.HasValue() {
		ctx = iIdentity.WithContext(ctx, coordinator.nodeIdentity)
	}

	getOpt := options.WithIdentity(options.GetDocument(), coordinator.nodeIdentity)
	doc, err := col.GetDocument(ctx, docIDType, getOpt)
	if err != nil {
		if errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) {
			return nil, nil
		}
		return nil, err
	}

	return generateDocArtifacts(ctx, col, doc, fieldNames, coordinator.nodeIdentity, coordinator.encKey)
}
