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
	"fmt"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	ipld "github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/storage/memstore"
	"github.com/sourcenetwork/corekv/memory"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	clientmocks "github.com/sourcenetwork/defradb/client/mocks"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
	message "github.com/sourcenetwork/defradb/internal/db/p2p/message"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/se/mocks"
)

// testSetup holds all test mocks and utilities for ReplicationCoordinator testing
type testSetup struct {
	t                *testing.T
	mockDB           *mocks.DB
	mockP2P          *mocks.P2P
	mockStorageProto *mockProto[PushSEArtifactsRequest, PushSEArtifactsReply]
	mockQueryProto   *mockProto[QuerySEArtifactsRequest, QuerySEArtifactsReply]
	mockEventBus     *mockEventBus
	coordinator      *ReplicationCoordinator
	rootstore        *memory.Datastore

	// Test data
	docID        string
	collectionID string
	fieldName    string
	peerID       string
	encKey       []byte
}

// newTestSetup creates a new test setup with all mocks initialized
func newTestSetup(t *testing.T) *testSetup {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	setup := &testSetup{
		t:                t,
		mockDB:           mocks.NewDB(t),
		mockP2P:          mocks.NewP2P(t),
		mockStorageProto: newMockProto[PushSEArtifactsRequest, PushSEArtifactsReply](t),
		mockQueryProto:   newMockProto[QuerySEArtifactsRequest, QuerySEArtifactsReply](t),
		mockEventBus: &mockEventBus{
			messages: make(chan event.Message, 10),
			subs:     make(map[event.Subscription]chan event.Message),
		},
		rootstore: rootstore,

		docID:        "bae-63c10140-a59a-5a7f-85d1-269e2c3841a6",
		collectionID: "test-collection",
		fieldName:    "age",
		peerID:       "peer1",
		encKey:       []byte("test-encryption-key-32-bytes-!"),
	}

	setup.mockDB.EXPECT().Events().Return(setup.mockEventBus).Maybe()
	setup.mockDB.EXPECT().MaxTxnRetries().Return(3).Maybe()
	setup.mockDB.EXPECT().Rootstore().Return(rootstore).Maybe()

	setup.createCoordinator()

	return setup
}

// createCoordinator creates the ReplicationCoordinator with all mocks
func (s *testSetup) createCoordinator() {
	rc, err := newReplicationCoordinator(
		s.mockDB,
		s.mockP2P,
		s.encKey,
		s.mockStorageProto,
		s.mockQueryProto,
	)
	require.NoError(s.t, err)
	s.coordinator = rc
}

// expectSEArtifactPush sets up expectation for SE artifact push to peer
func (s *testSetup) expectSEArtifactPush() {
	mockCollection := s.createMockCollectionWithDocument()

	s.mockDB.EXPECT().GetCollections(mock.Anything, mock.Anything).Return([]client.Collection{mockCollection}, nil)

	s.mockGetReplicatorsIDs([]string{s.peerID})

	// Expect SE artifact push to peer
	s.mockStorageProto.EXPECT().SendRequest(
		mock.Anything,
		mock.MatchedBy(func(req PushSEArtifactsRequest) bool {
			return req.CollectionID == s.collectionID && len(req.Artifacts) > 0
		}),
		s.peerID,
	).Return(PushSEArtifactsReply{}, nil)
}

// publishEvent publishes any event with the given name and data to the event bus
func (s *testSetup) publishEvent(name event.Name, evt any) {
	s.mockEventBus.Publish(event.NewMessage(name, evt))
}

// createValidCompositeBlock creates a proper CBOR-encoded composite block
func (s *testSetup) createValidCompositeBlock() []byte {
	return createValidCompositeBlock(s.t, s.docID, s.collectionID, s.fieldName)
}

// waitForArtifactPush waits for SE artifacts to be pushed and validates with custom assertion
func (s *testSetup) waitForArtifactPush(validate func(*PushSEArtifactsRequest) bool) {
	require.Eventually(s.t, func() bool {
		calls := s.mockStorageProto.Calls
		for _, call := range calls {
			if call.Method == "SendRequest" && len(call.Arguments) > 1 {
				if req, ok := call.Arguments[1].(PushSEArtifactsRequest); ok {
					return validate(&req)
				}
			}
		}
		return false
	}, time.Second, 10*time.Millisecond, "SE artifacts should be pushed to replicator with expected data")
}

// waitForNoCalls verifies that no calls were made to the storage protocol
func (s *testSetup) waitForNoCalls() {
	// Wait a bit to ensure no async calls happen
	time.Sleep(20 * time.Millisecond)
	require.Empty(s.t, s.mockStorageProto.Calls, "No SE artifacts should be pushed")
}

func (s *testSetup) mockGetReplicatorsIDs(peers []string) {
	s.mockP2P.EXPECT().GetReplicatorsIDs(s.collectionID).Return(peers).Maybe()
}

func (s *testSetup) mockGetCollections(m ...*clientmocks.Collection) {
	cols := make([]client.Collection, len(m))
	for i, col := range m {
		cols[i] = col
	}
	s.mockDB.EXPECT().GetCollections(mock.Anything, mock.Anything).Maybe().Return(cols, nil)
}

func getCollectionFieldsDescriptions() []client.CollectionFieldDescription {
	return []client.CollectionFieldDescription{
		{
			Name: "age",
			Kind: client.FieldKind_NILLABLE_INT,
		},
	}
}

func (s *testSetup) createEncryptedIndexesDescriptions() []client.EncryptedIndexDescription {
	return []client.EncryptedIndexDescription{
		{FieldName: s.fieldName, Type: client.EncryptedIndexTypeEquality},
	}
}

func (s *testSetup) createCollectionVersion() client.CollectionVersion {
	return client.CollectionVersion{
		Name:             "TestCollection",
		CollectionID:     s.collectionID,
		Fields:           getCollectionFieldsDescriptions(),
		EncryptedIndexes: s.createEncryptedIndexesDescriptions(),
	}
}

// createMockCollection creates a configurable mock collection
func (s *testSetup) createMockCollection() *clientmocks.Collection {
	mockCollection := clientmocks.NewCollection(s.t)

	mockCollection.EXPECT().Name().Return("TestCollection").Maybe()
	mockCollection.EXPECT().CollectionID().Return(s.collectionID).Maybe()
	mockCollection.EXPECT().VersionID().Return("v1").Maybe()

	mockCollection.EXPECT().GetEncryptedIndexes(mock.Anything).Return(
		[]client.EncryptedIndexDescription{
			{FieldName: s.fieldName, Type: client.EncryptedIndexTypeEquality},
		}, nil).Maybe()

	mockCollection.EXPECT().Version().Return(s.createCollectionVersion()).Maybe()

	return mockCollection
}

// createMockCollectionWithDocument creates a mock collection that returns a successful Get
func (s *testSetup) createMockCollectionWithDocument() *clientmocks.Collection {
	mockCollection := s.createMockCollection()

	// Setup Get method with default return
	mockCollection.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context, docID client.DocID, showDeleted bool) (*client.Document, error) {
			doc, err := client.NewDocFromMap(map[string]any{"age": 21}, mockCollection.Version())
			return doc, err
		}).Maybe()

	return mockCollection
}

// createNonCompositeBlock creates a non-composite block for testing
func (s *testSetup) createNonCompositeBlock() []byte {
	fieldBlock := coreblock.Block{
		Delta: crdt.CRDT{
			LWWDelta: &crdt.LWWDelta{
				DocID:           []byte(s.docID),
				FieldName:       s.fieldName,
				Priority:        1,
				SchemaVersionID: s.collectionID,
				Data:            []byte("21"),
			},
		},
	}

	blockBytes, err := fieldBlock.Marshal()
	require.NoError(s.t, err)
	return blockBytes
}

// close cleans up the coordinator and verifies all expectations
func (s *testSetup) close() {
	if s.coordinator != nil {
		s.coordinator.Close()
	}
	s.mockDB.AssertExpectations(s.t)
	s.mockP2P.AssertExpectations(s.t)
	s.mockStorageProto.AssertExpectations(s.t)
}

func (setup *testSetup) makeUpdateEvent() event.Update {
	updateEvent := event.Update{
		DocID:        setup.docID,
		CollectionID: setup.collectionID,
		Block:        setup.createValidCompositeBlock(),
	}
	return updateEvent
}

// createValidCompositeBlock creates a proper CBOR-encoded composite block using the pattern from block_test.go
func createValidCompositeBlock(t *testing.T, docID, collectionID, fieldName string) []byte {
	store := &memstore.Store{}
	lsys := cidlink.DefaultLinkSystem()
	lsys.SetReadStorage(store)
	lsys.SetWriteStorage(store)

	fieldBlock := coreblock.Block{
		Delta: crdt.CRDT{
			LWWDelta: &crdt.LWWDelta{
				DocID:           []byte(docID),
				FieldName:       fieldName,
				Priority:        1,
				SchemaVersionID: collectionID,
				Data:            []byte("21"),
			},
		},
	}
	fieldBlockLink, err := lsys.Store(ipld.LinkContext{}, coreblock.GetLinkPrototype(), fieldBlock.GenerateNode())
	require.NoError(t, err)

	compositeBlock := coreblock.Block{
		Delta: crdt.CRDT{
			DocCompositeDelta: &crdt.DocCompositeDelta{
				DocID:           []byte(docID),
				Priority:        1,
				SchemaVersionID: collectionID,
				Status:          1,
			},
		},
		Links: []coreblock.DAGLink{
			{
				Name: fieldName,
				Link: fieldBlockLink.(cidlink.Link),
			},
		},
	}

	blockBytes, err := compositeBlock.Marshal()
	require.NoError(t, err)
	return blockBytes
}

func TestReplicationCoordinator_WhenUpdateEventReceived_ShouldPushSEArtifactsToPeers(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	setup.expectSEArtifactPush()

	setup.publishEvent(event.UpdateName, setup.makeUpdateEvent())

	setup.waitForArtifactPush(func(req *PushSEArtifactsRequest) bool {
		return req.CollectionID == setup.collectionID && len(req.Artifacts) > 0
	})
}

func TestReplicationCoordinator_WhenBlockFailsToDeserialize_ShouldReturnError(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	updateEvent := event.Update{
		DocID:        setup.docID,
		CollectionID: setup.collectionID,
		Block:        []byte("invalid-block-data"),
	}
	setup.publishEvent(event.UpdateName, updateEvent)

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenNonCompositeBlock_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	updateEvent := event.Update{
		DocID:        setup.docID,
		CollectionID: setup.collectionID,
		Block:        setup.createNonCompositeBlock(),
	}
	setup.publishEvent(event.UpdateName, updateEvent)

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenGetCollectionsFails_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	setup.mockDB.EXPECT().GetCollections(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("database error"))

	setup.publishEvent(event.UpdateName, setup.makeUpdateEvent())

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenCollectionNotFound_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	setup.mockGetCollections()

	setup.publishEvent(event.UpdateName, setup.makeUpdateEvent())

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenInvalidDocID_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	setup.docID = "invalid-doc-id"
	setup.mockGetCollections(setup.createMockCollectionWithDocument())

	setup.publishEvent(event.UpdateName, setup.makeUpdateEvent())

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenDocumentNotFound_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	mockCollection := setup.createMockCollection()
	mockCollection.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, client.ErrDocumentNotFoundOrNotAuthorized).Maybe()
	setup.mockGetCollections(mockCollection)

	setup.mockGetReplicatorsIDs([]string{})

	setup.publishEvent(event.UpdateName, setup.makeUpdateEvent())

	require.Empty(setup.t, setup.mockStorageProto.Calls, "No SE artifacts should be pushed")
}

func TestReplicationCoordinator_WhenDocumentGetFails_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	mockCollection := setup.createMockCollection()
	mockCollection.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("database error")).Maybe()
	setup.mockGetCollections(mockCollection)

	setup.publishEvent(event.UpdateName, setup.makeUpdateEvent())

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenNoEncryptedIndexes_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	mockCollection := setup.createMockCollection()
	mockCollection.EXPECT().GetEncryptedIndexes(mock.Anything).Return(
		[]client.EncryptedIndexDescription{}, nil).Maybe()

	ver := setup.createCollectionVersion()
	ver.EncryptedIndexes = []client.EncryptedIndexDescription{}
	mockCollection.EXPECT().Version().Return(ver).Maybe()

	mockCollection.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context, docID client.DocID, showDeleted bool) (*client.Document, error) {
			doc, err := client.NewDocFromMap(map[string]any{"age": 21}, ver)
			return doc, err
		}).Maybe()

	setup.mockGetCollections(mockCollection)

	setup.mockGetReplicatorsIDs([]string{})

	setup.publishEvent(event.UpdateName, setup.makeUpdateEvent())

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenPushToReplicatorFails_ShouldStoreRetryInPeerstore(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	setup.mockGetCollections(setup.createMockCollectionWithDocument())

	setup.mockGetReplicatorsIDs([]string{setup.peerID})

	setup.mockStorageProto.EXPECT().SendRequest(mock.Anything, mock.Anything, setup.peerID).
		Return(PushSEArtifactsReply{}, fmt.Errorf("network error"))

	setup.publishEvent(event.UpdateName, setup.makeUpdateEvent())

	require.Eventually(t, func() bool {
		ps := datastore.PeerstoreFrom(setup.rootstore)
		retryKey := keys.NewPeerstoreSERetry(setup.peerID, setup.collectionID, setup.docID)
		has, err := ps.Has(context.Background(), retryKey.Bytes())
		require.NoError(t, err)

		if has {
			value, err := ps.Get(context.Background(), retryKey.Bytes())
			require.NoError(t, err)

			var retryInfo SERetryInfo
			err = cbor.Unmarshal(value, &retryInfo)
			require.NoError(t, err)

			require.Equal(t, setup.docID, retryInfo.DocID)
			require.Equal(t, setup.collectionID, retryInfo.CollectionID)
			require.Contains(t, retryInfo.FieldNames, setup.fieldName)
			require.Equal(t, 0, retryInfo.NumRetries)
			require.False(t, retryInfo.Retrying)
			require.NotZero(t, retryInfo.NextRetry)

			return true
		}
		return false
	}, time.Second, 10*time.Millisecond, "Retry data should be stored in peerstore after push failure")
}

func newMockProto[Req, Rep any](t *testing.T) *mockProto[Req, Rep] {
	return &mockProto[Req, Rep]{
		Mock: mock.Mock{},
		t:    t,
	}
}

type mockProto[Req, Rep any] struct {
	mock.Mock
	t testing.TB
}

func (m *mockProto[Req, Rep]) SendRequest(ctx context.Context, req Req, peerID string) (Rep, error) {
	args := m.Called(ctx, req, peerID)
	return args.Get(0).(Rep), args.Error(1)
}

func (m *mockProto[Req, Rep]) EXPECT() *mockProtoExpectation[Req, Rep] {
	return &mockProtoExpectation[Req, Rep]{mock: &m.Mock}
}

type mockProtoExpectation[Req, Rep any] struct {
	mock *mock.Mock
}

func (e *mockProtoExpectation[Req, Rep]) SendRequest(ctx, req, peerID interface{}) *mock.Call {
	return e.mock.On("SendRequest", ctx, req, peerID)
}

// Tests for processPushSEArtifactsRequest

func TestProcessPushSEArtifactsRequest_WhenArtifactsProvided_ShouldStoreInDatastore(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	req := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-123",
		},
		CollectionID: setup.collectionID,
		Artifacts: []SEArtifact{
			{
				DocID:     setup.docID,
				IndexID:   "index-1",
				SearchTag: []byte("search-tag-1"),
			},
		},
	}

	err := setup.coordinator.processPushSEArtifactsRequest(context.Background(), req)
	require.NoError(t, err)

	ds := datastore.DatastoreFrom(setup.rootstore)
	key := keys.DatastoreSE{
		CollectionID: setup.collectionID,
		IndexID:      "index-1",
		SearchTag:    []byte("search-tag-1"),
		DocID:        setup.docID,
	}

	has, err := ds.Has(context.Background(), key.Bytes())
	require.NoError(t, err)
	require.True(t, has, "Artifact should be stored in datastore")
}

func TestProcessPushSEArtifactsRequest_WhenMultipleArtifacts_ShouldStoreAll(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	req := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-123",
		},
		CollectionID: setup.collectionID,
		Artifacts: []SEArtifact{
			{
				DocID:     setup.docID,
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
			{
				DocID:     "doc-2",
				IndexID:   "index-2",
				SearchTag: []byte("tag-2"),
			},
			{
				DocID:     "doc-3",
				IndexID:   "index-3",
				SearchTag: []byte("tag-3"),
			},
		},
	}

	err := setup.coordinator.processPushSEArtifactsRequest(context.Background(), req)
	require.NoError(t, err)

	ds := datastore.DatastoreFrom(setup.rootstore)

	for i, artifact := range req.Artifacts {
		key := keys.DatastoreSE{
			CollectionID: setup.collectionID,
			IndexID:      artifact.IndexID,
			SearchTag:    artifact.SearchTag,
			DocID:        artifact.DocID,
		}

		has, err := ds.Has(context.Background(), key.Bytes())
		require.NoError(t, err)
		require.True(t, has, "Artifact %d should be stored in datastore", i+1)
	}
}

func TestProcessPushSEArtifactsRequest_WhenEmptyArtifacts_ShouldSucceed(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	req := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-123",
		},
		CollectionID: setup.collectionID,
		Artifacts:    []SEArtifact{},
	}

	err := setup.coordinator.processPushSEArtifactsRequest(context.Background(), req)
	require.NoError(t, err, "Should succeed with empty artifacts list")
}

func TestProcessPushSEArtifactsRequest_WhenDuplicateArtifacts_ShouldOverwrite(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	artifact := SEArtifact{
		DocID:     setup.docID,
		IndexID:   "index-1",
		SearchTag: []byte("search-tag"),
	}

	req1 := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-1",
		},
		CollectionID: setup.collectionID,
		Artifacts:    []SEArtifact{artifact},
	}

	err := setup.coordinator.processPushSEArtifactsRequest(context.Background(), req1)
	require.NoError(t, err)

	req2 := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-2",
		},
		CollectionID: setup.collectionID,
		Artifacts:    []SEArtifact{artifact},
	}

	err = setup.coordinator.processPushSEArtifactsRequest(context.Background(), req2)
	require.NoError(t, err, "Should succeed when storing duplicate artifact")

	ds := datastore.DatastoreFrom(setup.rootstore)
	key := keys.DatastoreSE{
		CollectionID: setup.collectionID,
		IndexID:      artifact.IndexID,
		SearchTag:    artifact.SearchTag,
		DocID:        artifact.DocID,
	}

	has, err := ds.Has(context.Background(), key.Bytes())
	require.NoError(t, err)
	require.True(t, has, "Artifact should exist in datastore")
}

// Tests for processQuerySEArtifactsRequest

func TestProcessQuerySEArtifactsRequest_WhenMatchingArtifacts_ShouldReturnDocIDs(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	artifacts := []SEArtifact{
		{
			DocID:     "doc-1",
			IndexID:   "index-1",
			SearchTag: []byte("tag-1"),
		},
		{
			DocID:     "doc-2",
			IndexID:   "index-1",
			SearchTag: []byte("tag-1"),
		},
		{
			DocID:     "doc-3",
			IndexID:   "index-2",
			SearchTag: []byte("tag-2"),
		},
	}

	storeReq := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-setup",
		},
		CollectionID: setup.collectionID,
		Artifacts:    artifacts,
	}
	err := setup.coordinator.processPushSEArtifactsRequest(context.Background(), storeReq)
	require.NoError(t, err)

	queryReq := &QuerySEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-query",
		},
		CollectionID: setup.collectionID,
		Queries: []SEFieldQuery{
			{
				FieldName: "field1",
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
		},
	}

	reply, err := setup.coordinator.processQuerySEArtifactsRequest(context.Background(), queryReq)
	require.NoError(t, err)
	require.NotNil(t, reply.DocIDs)
	require.Len(t, reply.DocIDs, 2, "Should return 2 matching documents")
	require.Contains(t, reply.DocIDs, "doc-1")
	require.Contains(t, reply.DocIDs, "doc-2")
}

func TestProcessQuerySEArtifactsRequest_WhenNoMatchingArtifacts_ShouldReturnEmpty(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	storeReq := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-setup",
		},
		CollectionID: setup.collectionID,
		Artifacts: []SEArtifact{
			{
				DocID:     "doc-1",
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
		},
	}
	err := setup.coordinator.processPushSEArtifactsRequest(context.Background(), storeReq)
	require.NoError(t, err)

	queryReq := &QuerySEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-query",
		},
		CollectionID: setup.collectionID,
		Queries: []SEFieldQuery{
			{
				FieldName: "field1",
				IndexID:   "non-existent-index",
				SearchTag: []byte("non-existent-tag"),
			},
		},
	}

	reply, err := setup.coordinator.processQuerySEArtifactsRequest(context.Background(), queryReq)
	require.NoError(t, err)
	require.Empty(t, reply.DocIDs, "Should return empty list when no matches found")
}

func TestProcessQuerySEArtifactsRequest_WhenMultipleQueries_ShouldReturnUnion(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	artifacts := []SEArtifact{
		{
			DocID:     "doc-1",
			IndexID:   "index-1",
			SearchTag: []byte("tag-1"),
		},
		{
			DocID:     "doc-2",
			IndexID:   "index-2",
			SearchTag: []byte("tag-2"),
		},
		{
			DocID:     "doc-3",
			IndexID:   "index-3",
			SearchTag: []byte("tag-3"),
		},
	}

	storeReq := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-setup",
		},
		CollectionID: setup.collectionID,
		Artifacts:    artifacts,
	}
	err := setup.coordinator.processPushSEArtifactsRequest(context.Background(), storeReq)
	require.NoError(t, err)

	queryReq := &QuerySEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-query",
		},
		CollectionID: setup.collectionID,
		Queries: []SEFieldQuery{
			{
				FieldName: "field1",
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
			{
				FieldName: "field2",
				IndexID:   "index-2",
				SearchTag: []byte("tag-2"),
			},
		},
	}

	reply, err := setup.coordinator.processQuerySEArtifactsRequest(context.Background(), queryReq)
	require.NoError(t, err)
	require.Len(t, reply.DocIDs, 2, "Should return union of matching documents")
	require.Contains(t, reply.DocIDs, "doc-1")
	require.Contains(t, reply.DocIDs, "doc-2")
}

func TestProcessQuerySEArtifactsRequest_WhenEmptyQueries_ShouldReturnEmpty(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	queryReq := &QuerySEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-query",
		},
		CollectionID: setup.collectionID,
		Queries:      []SEFieldQuery{},
	}

	reply, err := setup.coordinator.processQuerySEArtifactsRequest(context.Background(), queryReq)
	require.NoError(t, err)
	require.Empty(t, reply.DocIDs, "Should return empty list for empty queries")
}

func TestProcessQuerySEArtifactsRequest_WhenDifferentCollections_ShouldOnlyReturnFromSpecified(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	artifact1 := SEArtifact{
		DocID:     "doc-1",
		IndexID:   "index-1",
		SearchTag: []byte("tag-1"),
	}

	storeReq1 := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-setup",
		},
		CollectionID: setup.collectionID,
		Artifacts:    []SEArtifact{artifact1},
	}
	err := setup.coordinator.processPushSEArtifactsRequest(context.Background(), storeReq1)
	require.NoError(t, err)

	artifact2 := SEArtifact{
		DocID:     "doc-2",
		IndexID:   "index-1",
		SearchTag: []byte("tag-1"),
	}

	storeReq2 := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-setup",
		},
		CollectionID: "other-collection",
		Artifacts:    []SEArtifact{artifact2},
	}
	err = setup.coordinator.processPushSEArtifactsRequest(context.Background(), storeReq2)
	require.NoError(t, err)

	queryReq := &QuerySEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-query",
		},
		CollectionID: setup.collectionID,
		Queries: []SEFieldQuery{
			{
				FieldName: "field1",
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
		},
	}

	reply, err := setup.coordinator.processQuerySEArtifactsRequest(context.Background(), queryReq)
	require.NoError(t, err)
	require.Len(t, reply.DocIDs, 1, "Should only return documents from specified collection")
	require.Contains(t, reply.DocIDs, "doc-1")
	require.NotContains(t, reply.DocIDs, "doc-2", "Should not return documents from other collections")
}

// Tests for handleQuerySEArtifactsEvent

func TestHandleQuerySEArtifactsEvent_WhenReplicatorsExist_ShouldQueryAndReturnDocIDs(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	responseChan := make(chan SEArtifactsResult, 1)
	evt := RequestSEArtifactsEvent{
		CollectionID: setup.collectionID,
		Queries: []FieldQuery{
			{
				FieldName: "field1",
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
		},
		Response: responseChan,
	}

	setup.mockGetReplicatorsIDs([]string{setup.peerID})

	expectedReply := QuerySEArtifactsReply{DocIDs: []string{"doc-1", "doc-2"}}
	setup.mockQueryProto.EXPECT().SendRequest(mock.Anything, mock.Anything, setup.peerID).Return(expectedReply, nil)

	setup.coordinator.handleQuerySEArtifactsEvent(evt)

	result := <-responseChan
	require.NoError(t, result.Error)
	require.Len(t, result.DocIDs, 2)
	require.Contains(t, result.DocIDs, "doc-1")
	require.Contains(t, result.DocIDs, "doc-2")
}

func TestHandleQuerySEArtifactsEvent_WhenNoReplicators_ShouldReturnEmpty(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	responseChan := make(chan SEArtifactsResult, 1)
	evt := RequestSEArtifactsEvent{
		CollectionID: setup.collectionID,
		Queries: []FieldQuery{
			{
				FieldName: "field1",
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
		},
		Response: responseChan,
	}

	setup.mockGetReplicatorsIDs([]string{})

	setup.coordinator.handleQuerySEArtifactsEvent(evt)

	result := <-responseChan
	require.NoError(t, result.Error)
	require.Empty(t, result.DocIDs)
}

func TestHandleQuerySEArtifactsEvent_WhenFirstReplicatorFails_ShouldTryNext(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	responseChan := make(chan SEArtifactsResult, 1)
	evt := RequestSEArtifactsEvent{
		CollectionID: setup.collectionID,
		Queries: []FieldQuery{
			{
				FieldName: "field1",
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
		},
		Response: responseChan,
	}

	peerID1 := "peer-1"
	peerID2 := "peer-2"
	setup.mockGetReplicatorsIDs([]string{peerID1, peerID2})

	setup.mockQueryProto.EXPECT().SendRequest(mock.Anything, mock.Anything, peerID1).
		Return(QuerySEArtifactsReply{}, fmt.Errorf("network error")).Once()

	expectedReply := QuerySEArtifactsReply{
		DocIDs: []string{"doc-3", "doc-4"},
	}
	setup.mockQueryProto.EXPECT().SendRequest(mock.Anything, mock.Anything, peerID2).Return(expectedReply, nil).Once()

	setup.coordinator.handleQuerySEArtifactsEvent(evt)

	result := <-responseChan
	require.NoError(t, result.Error)
	require.Len(t, result.DocIDs, 2)
	require.Contains(t, result.DocIDs, "doc-3")
	require.Contains(t, result.DocIDs, "doc-4")
}

func TestHandleQuerySEArtifactsEvent_WhenAllReplicatorsFail_ShouldReturnError(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	responseChan := make(chan SEArtifactsResult, 1)
	evt := RequestSEArtifactsEvent{
		CollectionID: setup.collectionID,
		Queries: []FieldQuery{
			{
				FieldName: "field1",
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
		},
		Response: responseChan,
	}

	peerID1 := "peer-1"
	peerID2 := "peer-2"
	setup.mockGetReplicatorsIDs([]string{peerID1, peerID2})

	setup.mockQueryProto.EXPECT().SendRequest(mock.Anything, mock.Anything, peerID1).
		Return(QuerySEArtifactsReply{}, fmt.Errorf("network error 1")).Once()

	setup.mockQueryProto.EXPECT().SendRequest(mock.Anything, mock.Anything, peerID2).
		Return(QuerySEArtifactsReply{}, fmt.Errorf("network error 2")).Once()

	setup.coordinator.handleQuerySEArtifactsEvent(evt)

	result := <-responseChan
	require.Error(t, result.Error)
	require.Empty(t, result.DocIDs)
	require.Contains(t, result.Error.Error(), "network error 2")
}

func TestHandleQuerySEArtifactsEvent_WhenMultipleQueries_ShouldPassAllToReplicator(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	responseChan := make(chan SEArtifactsResult, 1)
	evt := RequestSEArtifactsEvent{
		CollectionID: setup.collectionID,
		Queries: []FieldQuery{
			{
				FieldName: "field1",
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
			{
				FieldName: "field2",
				IndexID:   "index-2",
				SearchTag: []byte("tag-2"),
			},
			{
				FieldName: "field3",
				IndexID:   "index-3",
				SearchTag: []byte("tag-3"),
			},
		},
		Response: responseChan,
	}

	setup.mockGetReplicatorsIDs([]string{setup.peerID})

	expectedReply := QuerySEArtifactsReply{
		DocIDs: []string{"doc-1", "doc-2", "doc-3"},
	}
	setup.mockQueryProto.EXPECT().SendRequest(
		mock.Anything,
		mock.MatchedBy(func(req QuerySEArtifactsRequest) bool {
			return req.CollectionID == setup.collectionID && len(req.Queries) == 3
		}),
		setup.peerID,
	).Return(expectedReply, nil)

	setup.coordinator.handleQuerySEArtifactsEvent(evt)

	result := <-responseChan
	require.NoError(t, result.Error)
	require.Len(t, result.DocIDs, 3)
	require.Contains(t, result.DocIDs, "doc-1")
	require.Contains(t, result.DocIDs, "doc-2")
	require.Contains(t, result.DocIDs, "doc-3")
}

type mockEventBus struct {
	messages chan event.Message
	subs     map[event.Subscription]chan event.Message
}

func (m *mockEventBus) Publish(msg event.Message) {
	for _, ch := range m.subs {
		select {
		case ch <- msg:
		default:
			// Don't block if channel is full
		}
	}
}

func (m *mockEventBus) Subscribe(events ...event.Name) (event.Subscription, error) {
	ch := make(chan event.Message, 10)
	sub := &mockSubscription{ch: ch}
	m.subs[sub] = ch
	return sub, nil
}

func (m *mockEventBus) Unsubscribe(sub event.Subscription) {
	if ch, exists := m.subs[sub]; exists {
		close(ch)
		delete(m.subs, sub)
	}
}

func (m *mockEventBus) Close() {
	for _, ch := range m.subs {
		close(ch)
	}
}

// mockSubscription implements event.Subscription for testing
type mockSubscription struct {
	ch chan event.Message
}

func (m *mockSubscription) Message() <-chan event.Message {
	return m.ch
}

// Removed mockSimpleCollection as we're now using the generated mocks from client/mocks
